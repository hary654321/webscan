package node

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"
	"webscan/config"
	"webscan/internal/domain/entity"
	"webscan/internal/domain/repository/node"
	"webscan/internal/infra/cache"
	"webscan/internal/infra/persistence"
	"webscan/pkg/crawlergo"
	"webscan/pkg/nuclei"
	"webscan/utils/log"
)

type Repository struct {
	config config.ConfigureInterface
	logger log.LoggerInterface
	//爬虫
	spider crawlergo.SpiderInterface
	//扫描器
	nuclei nuclei.ScannerInterface
	//数据库
	mySqlite *gorm.DB
	//内存缓存
	stateCache cache.StateCacheInterface
}

var _ node.RepositoryInterface = &Repository{}

func NewNodePersistence(
	config config.ConfigureInterface,
	logger log.LoggerInterface,
	spider crawlergo.SpiderInterface,
	nuclei nuclei.ScannerInterface,
	stateCache cache.StateCacheInterface,
) *Repository {
	// 扫描节点需要连接 本地sqlite数据库
	_, mySqlite := persistence.InitDatabase(config)
	// NOTE: 自动创建表
	mySqlite.AutoMigrate(&entity.NodeTask{})

	return &Repository{
		config:     config,
		logger:     logger,
		spider:     spider,
		nuclei:     nuclei,
		mySqlite:   mySqlite,
		stateCache: stateCache,
	}
}

// NewTask 创建新任务记录
func (r *Repository) NewTask(task *entity.NodeTask) (uint, error) {
	tx := r.mySqlite.Create(task)
	return task.ID, tx.Error
}

// StartTask 启动任务：启动爬虫爬取目标 启动Nuclei扫描目标
func (r *Repository) StartTask(task *entity.NodeTask) error {
	// 在缓存中插入任务状态
	r.stateCache.New(task.ID)
	// 更新任务状态为running
	err := r.mySqlite.Model(&entity.NodeTask{}).Where("id = ?", task.ID).Update("status", "crawling").Error
	if err != nil {
		r.logger.Errorf("[StartTask] task id: %d, target id: %d, update status error: %v", task.TaskID, task.TargetID, err)
		r.stateCache.Delete(task.ID)
		return err
	}
	// 启动爬虫
	spiderResult, err := r.StartSpider(task)
	if err != nil {
		r.logger.Errorf("[StartTask] task id: %d, target id: %d, spider error: %v",
			task.TaskID, task.TargetID, err)
		// 如果是主动退出则不修改数据库状态
		if err.Error() == "received_stop_signal" {
			return nil
		}
		// 爬虫执行失败
		updates := map[string]interface{}{
			"status":       "error",
			"error_reason": err.Error(),
		}
		err = r.mySqlite.Model(&entity.NodeTask{}).Where("id = ?", task.ID).Updates(updates).Error
		if err != nil {
			r.logger.Errorf("[StartTask] task id: %d, target id: %d, update status error: %v", task.TaskID, task.TargetID, err)
		}
		r.stateCache.Delete(task.ID)

		return err
	}
	// 将爬虫执行结果写到文件里
	r.writeSpiderResultFile(task, spiderResult)

	//sr := make(map[string][]string)
	//sr["req_list"] = spiderResult.ReqList
	//sr["all_req_list"] = spiderResult.AllReqList
	//sr["all_domain_list"] = spiderResult.AllDomainList
	//sr["sub_domain_list"] = spiderResult.SubDomainList
	// 更新表状态、缓存状态
	r.stateCache.UpdateSpiderFinishedState(task.ID, spiderResult.ReqList, spiderResult.AllReqList, spiderResult.AllDomainList, spiderResult.SubDomainList)

	r.mySqlite.Model(&entity.NodeTask{}).Where("id = ?", task.ID).
		Where("status != ?", "paused").
		Update("status", "crawled")

	// 启动扫描器进行扫描
	err = r.StartNuclei(task)
	if err != nil {
		r.logger.Errorf("[StartTask] task id: %d, target id: %d, start scanner error: %v",
			task.TaskID, task.TargetID, err)
		// 如果是主动退出则不修改数据库状态
		if err.Error() == "received_stop_signal" {
			return nil
		}
		// 扫描器执行失败
		updates := map[string]interface{}{
			"status":       "error",
			"error_reason": err.Error(),
		}
		err = r.mySqlite.Model(&entity.NodeTask{}).Where("id = ?", task.ID).Updates(updates).Error
		if err != nil {
			r.logger.Errorf("[StartTask] task id: %d, target id: %d, update status error: %v", task.TaskID, task.TargetID, err)
		}
		r.stateCache.Delete(task.ID)

		return err
	}
	// 任务执行结束，更新任务状态
	r.mySqlite.Model(&entity.NodeTask{}).Where("id = ?", task.ID).Update("status", "finished")
	r.stateCache.UpdateScannerState(task.ID, "finished")
	return nil
}

// StartSpider 启动爬虫爬虫目标
func (r *Repository) StartSpider(task *entity.NodeTask) (crawlergo.SpiderResult, error) {
	var spiderResult crawlergo.SpiderResult
	// 解码爬虫配置
	var spiderConfig *entity.Crawlergo
	spiderConfig = &entity.Crawlergo{}
	err := json.Unmarshal([]byte(task.SpiderConfig), spiderConfig)
	if err != nil {
		return spiderResult, err
	}

	spiderFn := func(start, stop chan struct{}, target string, spiderConfig *entity.Crawlergo) (crawlergo.SpiderResult, error) {
		var result crawlergo.SpiderResult
		var err error
		for {
			select {
			case _, ok := <-stop:
				if !ok {
					r.logger.Infof("[StartSpider] task id: %d, target id: %d, stop signal channel closed", task.TaskID, task.TargetID)
				}
				// 退出信号 Received a stop signal
				r.logger.Errorf("[StartSpider] task id: %d, target id: %d, received stop signal", task.TaskID, task.TargetID)
				err = errors.New("received_stop_signal")
				goto gout
			case <-start:
				// do work
				result, err = r.spider.Start(target, spiderConfig)
				goto gout
			}
		}

	gout:
		return result, err
	}

	spiderStartChan := make(chan struct{}, 1)
	spiderStartChan <- struct{}{}
	defer close(spiderStartChan)

	if r.stateCache.Get(task.ID) == nil {
		return spiderResult, errors.New("received_stop_signal")
	}

	spiderResult, err = spiderFn(spiderStartChan, r.stateCache.Get(task.ID).SpiderStopChan, task.Target, spiderConfig)
	return spiderResult, err
}

func (r *Repository) writeSpiderResultFile(task *entity.NodeTask, result crawlergo.SpiderResult) error {
	taskConfig := r.config.GetTaskConfig()
	// 爬虫执行结束，将结果写入缓存与文件
	// dir/1-1-1.crawl
	idStr := strconv.FormatUint(uint64(task.ID), 10)
	taskIDStr := strconv.FormatUint(uint64(task.TaskID), 10)
	targetIDStr := strconv.FormatUint(uint64(task.TargetID), 10)
	spiderOutputPath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl")

	spiderResultMarshal, err := json.Marshal(result)
	if err != nil {
		return err
	}
	f, err := os.Create(spiderOutputPath)
	if err != nil {
		return err
	}
	f.Write(spiderResultMarshal)
	f.Close()

	// scan file path: dir/1-1.crawl.discovered
	discoveredHostFilePath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl.discovered")
	discoveredHostFile, err := os.Create(discoveredHostFilePath)
	if err != nil {
		return err
	}
	defer discoveredHostFile.Close()

	for _, v := range result.AllDomainList {
		discoveredHostFile.WriteString(v + "\n")
	}

	// scan file path: dir/1-1.crawl.scan
	scanURL := make(map[string]struct{})
	//targetParse, _ := url.Parse(task.Target)
	for _, v := range result.ReqList {
		u, parseErr := url.Parse(v)
		if parseErr != nil {
			continue
		}

		if u.Host == "" {
			continue
		}

		//if path.Dir(targetParse.Path) != path.Dir(u.Path) {
		//	continue
		//}

		scanURL[v] = struct{}{}
	}

	scanFilePath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl.scan")
	scanFile, err := os.Create(scanFilePath)
	if err != nil {
		return err
	}
	defer scanFile.Close()

	for k, _ := range scanURL {
		scanFile.WriteString(k + "\n")
	}
	return nil
}

// StartNuclei 启动Nuclei扫描目标
func (r *Repository) StartNuclei(task *entity.NodeTask) error {
	scannerFn := func(start, stop chan struct{}) error {
		var err error
		for {
			select {
			case _, ok := <-stop:
				if !ok {
					r.logger.Infof("[StartNuclei] task id: %d, target id: %d, stop signal channel closed", task.TaskID, task.TargetID)
				}
				// 退出信号 Received a stop signal
				r.logger.Errorf("[StartNuclei] task id: %d, target id: %d, received stop signal", task.TaskID, task.TargetID)
				err = errors.New("received_stop_signal")
				goto gout
			case <-start:
				// do work
				idStr := strconv.FormatUint(uint64(task.ID), 10)
				taskIDStr := strconv.FormatUint(uint64(task.TaskID), 10)
				targetIDStr := strconv.FormatUint(uint64(task.TargetID), 10)

				scanFilePath := path.Join(r.config.GetTaskConfig().DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl.scan")
				var scans []string
				//检查文件是否存在
				if _, err := os.Stat(scanFilePath); os.IsNotExist(err) {
					scans = append(scans, task.Target)
				} else {
					// 打开文件
					file, err := os.Open(scanFilePath)
					if err != nil {
						return err
					}
					defer file.Close()

					// 按行读取文件内容
					scanner := bufio.NewScanner(file)
					// 逐行读取文件内容
					for scanner.Scan() {
						line := scanner.Text()
						scans = append(scans, line)
					}
					// 检查是否出现了扫描错误
					if err := scanner.Err(); err != nil {
						return err
					}
				}

				if len(scans) == 0 {
					return nil
				}

				taskConfig := r.config.GetTaskConfig()
				nucleiOutputPath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".scan")

				//获取模板目录
				templatesDir := r.config.GetAppConfig().Templates
				r.logger.Infof("[StartNuclei] task id: %d, target id: %d, templatesDir: %s", task.TaskID, task.TargetID, templatesDir)
				err = r.nuclei.StartV2(task, templatesDir, []string{scans[0]}, nucleiOutputPath)
				goto gout
			}
		}

	gout:
		return err
	}

	scannerStartChan := make(chan struct{}, 1)
	scannerStartChan <- struct{}{}
	defer close(scannerStartChan)

	if r.stateCache.Get(task.ID) == nil {
		return errors.New("received_stop_signal")
	}
	err := scannerFn(scannerStartChan, r.stateCache.Get(task.ID).ScannerStopChan)
	return err
}

func (r *Repository) StopTask(task *entity.NodeTask) error {
	// stop 需要直接结束任务，下次重新开始一个任务时，会在本地重新创建一条任务。
	err := r.mySqlite.Model(&entity.NodeTask{}).Where("id = ?", task.ID).Update("status", "stop").Error
	if err != nil {
		// 关闭爬虫channel/扫描器channel以停止任务，然后删除内存缓存
		r.stateCache.Delete(task.ID)
		return nil
	}
	return err
}

func (r *Repository) DeleteTasks(tasks []entity.NodeTask) {
	// 结束正在执行的任务，删除表记录，删除文件
	for _, task := range tasks {
		r.stateCache.Delete(task.ID)
		r.mySqlite.Where("id = ?", task.ID).Delete(&entity.NodeTask{})
		idStr := strconv.FormatUint(uint64(task.ID), 10)
		taskIDStr := strconv.FormatUint(uint64(task.TaskID), 10)
		targetIDStr := strconv.FormatUint(uint64(task.TargetID), 10)
		// 删除文件
		os.Remove(path.Join(r.config.GetTaskConfig().DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl"))
		os.Remove(path.Join(r.config.GetTaskConfig().DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl.discovered"))
		os.Remove(path.Join(r.config.GetTaskConfig().DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl.scan"))
		os.Remove(path.Join(r.config.GetTaskConfig().DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".scan"))
	}
}

func (r *Repository) PauseTask(task *entity.NodeTask) error {
	// 更新任务状态为pause
	r.mySqlite.Model(&entity.NodeTask{}).Where("id = ?", task.ID).Update("status", "paused")
	r.stateCache.Delete(task.ID)
	return nil
}

func (r *Repository) ResumeTask(task *entity.NodeTask) error {
	// TODO
	return nil
}

//------------

//// GetTaskByTaskAndTargetID 使用taskID查询多条记录
//func (r *Repository) GetTaskByTaskAndTargetID(taskID, targetID uint) (entity.NodeTask, error) {
//	var task entity.NodeTask
//	err := r.mySqlite.Where("task_id = ? and target_id = ?", taskID, targetID).First(&task).Error
//	return task, err
//}

//func (r *Repository) UpdateTaskStatus(id uint, key, value string) error {
//	updateFields := map[string]interface{}{
//		key: value,
//	}
//	err := r.mySqlite.Model(&entity.NodeTask{}).Where("id = ?", id).Updates(updateFields).Error
//	return err
//}

//type DiscoveredHost struct {
//	TaskID         uint     `json:"task_id"`
//	TargetID       uint     `json:"target_id"`
//	ScanTarget     string   `json:"scan_target"`
//	DiscoveredHost []string `json:"discovered_host"`
//}
//
//type CrawlURL struct {
//	TaskID     uint     `json:"task_id"`
//	TargetID   uint     `json:"target_id"`
//	ScanTarget string   `json:"scan_target"`
//	CrawlURL   []string `json:"craw_url"`
//}

//// StartNuclei 启动Nuclei扫描目标
//func (n *NodeRepository) StartNuclei(task *entity.NodeTask) error {
//	idStr := strconv.FormatUint(uint64(task.ID), 10)
//	taskIDStr := strconv.FormatUint(uint64(task.TaskID), 10)
//	targetIDStr := strconv.FormatUint(uint64(task.TargetID), 10)
//
//	scanFilePath := path.Join(n.config.GetTaskConfig().DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl.scan")
//	var scans []string
//	//检查文件是否存在
//	if _, err := on.Stat(scanFilePath); on.IsNotExist(err) {
//		scans = append(scans, task.Target)
//	} else {
//		// 打开文件
//		file, err := on.Open(scanFilePath)
//		if err != nil {
//			return err
//		}
//		defer file.Close()
//
//		// 按行读取文件内容
//		scanner := bufio.NewScanner(file)
//		// 逐行读取文件内容
//		for scanner.Scan() {
//			line := scanner.Text()
//			scans = append(scans, line)
//		}
//		// 检查是否出现了扫描错误
//		if err := scanner.Err(); err != nil {
//			return err
//		}
//	}
//
//	if len(scans) == 0 {
//		return nil
//	}
//
//	taskConfig := n.config.GetTaskConfig()
//	nucleiOutputPath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".scan")
//
//	//获取模板目录
//	templatesDir := n.config.GetAppConfig().Templates
//	n.logger.Infof("[StartNuclei] task id: %d, target id: %d, templatesDir: %s", task.TaskID, task.TargetID, templatesDir)
//	return n.nuclei.Start(task, templatesDir, []string{scans[0]}, nucleiOutputPath)
//}
//
//// GetNextTask 获取下一个要执行的任务
//func (n *NodeRepository) GetNextTask() (entity.NodeTask, error) {
//	var task entity.NodeTask
//	err := n.mySqlite.Where("status != ?", "finished").First(&task).Error
//	return task, err
//}

//func (r *Repository) GetTaskScanResult(taskID, targetID uint) (map[string]interface{}, error) {
//	var task entity.NodeTask
//	err := r.mySqlite.Where("task_id = ? and target_id = ?", taskID, targetID).Find(&task).Error
//	if err != nil {
//		return nil, err
//	}
//
//	taskConfig := r.configFace.GetTaskConfig()
//	idStr := strconv.FormatUint(uint64(task.ID), 10)
//	taskIDStr := strconv.FormatUint(uint64(task.TaskID), 10)
//	targetIDStr := strconv.FormatUint(uint64(task.TargetID), 10)
//
//	discoveredHostFilePath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl.discovered")
//	crawlFilePath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl.scan")
//	scanFilePath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".scan")
//
//	discoveredHostFile, err := os.Open(discoveredHostFilePath)
//	if err != nil {
//		return nil, err
//	}
//	defer discoveredHostFile.Close()
//	discoveredHostLines := readLines(discoveredHostFile)
//	dh := DiscoveredHost{
//		TaskID:         task.TaskID,
//		TargetID:       task.TargetID,
//		ScanTarget:     task.Target,
//		DiscoveredHost: discoveredHostLines,
//	}
//
//	crawlFile, err := os.Open(crawlFilePath)
//	if err != nil {
//		return nil, err
//	}
//	defer crawlFile.Close()
//	crawlLines := readLines(crawlFile)
//	cls := CrawlURL{
//		TaskID:     task.TaskID,
//		TargetID:   task.TargetID,
//		ScanTarget: task.Target,
//		CrawlURL:   crawlLines,
//	}
//
//	scanFile, err := os.Open(scanFilePath)
//	if err != nil {
//		return nil, err
//	}
//	defer scanFile.Close()
//	scanLines := readLines(scanFile)
//
//	var newScanLines []map[string]interface{}
//	for _, line := range scanLines {
//		var l map[string]interface{}
//		err := json.Unmarshal([]byte(line), &l)
//		if err != nil {
//			r.logFace.Errorf("json unmarshal error: %v", err)
//			continue
//		}
//
//		l["task_id"] = task.TaskID
//		l["target_id"] = task.TargetID
//		l["scan_target"] = task.Target
//		l["target_vuln_id"] = uuid.New().String()
//
//		newScanLines = append(newScanLines, l)
//	}
//
//	scanResults := make(map[string]interface{})
//	scanResults["discovered_host"] = dh
//	scanResults["crawl"] = cls
//	scanResults["scan"] = newScanLines
//
//	return scanResults, err
//}
//
//func (r *Repository) GetTaskScanStatus(taskID, targetID uint) (string, error) {
//	var task entity.NodeTask
//	err := r.mySqlite.Where("task_id = ? and target_id = ?", taskID, targetID).Find(&task).Error
//	return task.Status, err
//}

//func readLines(f io.Reader) []string {
//	scanner := bufio.NewScanner(f)
//
//	var lines []string
//	for scanner.Scan() {
//		line := scanner.Text()
//		lines = append(lines, line)
//	}
//
//	//// 检查扫描过程中是否出现错误
//	//if err := scanner.Err(); err != nil {
//	//
//	//}
//	return lines
//}

//// GetNextTask 获取下一个要执行的任务
//func (r *Repository) GetNextTask() (entity.NodeTask, error) {
//	var task entity.NodeTask
//	err := r.mySqlite.Where("status != ?", "finished").First(&task).Error
//	return task, err
//}

//// / ----------------
//func (r *Repository) TaskState(task *entity.NodeTask) ([]byte, error) {
//	state := r.stateCache.GetState(task.ID)
//	if state != nil {
//		stateByte, err := json.Marshal(state)
//		return stateByte, err
//	}
//	return nil, "not_found"
//}

func (r *Repository) FindOneTask(taskID, targetID uint) (entity.NodeTask, error) {
	var task entity.NodeTask
	err := r.mySqlite.Where("task_id = ? and target_id = ?", taskID, targetID).First(&task).Error
	return task, err
}

func (r *Repository) FindTask(taskID uint) ([]entity.NodeTask, error) {
	var task []entity.NodeTask
	err := r.mySqlite.Where("task_id = ?", taskID).Find(&task).Error
	return task, err
}

func (r *Repository) FindRunningTask(taskID uint) ([]entity.NodeTask, error) {
	var runningTask []entity.NodeTask
	err := r.mySqlite.Where("task_id = ?", taskID).
		Where("status IN (?)", []string{"crawling", "scanning"}).
		Find(&runningTask).Error
	if err != nil {
		return nil, err
	}

	return runningTask, err
}

// ResumePausedTask 恢复暂停的任务
func (r *Repository) ResumePausedTask(taskID uint) error {
	return r.mySqlite.Model(&entity.NodeTask{}).
		Where("task_id = ?", taskID).
		Where("status = ?", "paused").
		Update("status", "created").Error
}

// GetCrawledResult 获取爬虫结果
func (r *Repository) GetCrawledResult(id, taskID, targetID uint, target string) (map[string]interface{}, error) {
	//var task entity.SlaveTask
	//err := s.mySqlite.Where("task_id = ? and target_id = ?", taskID, targetID).Find(&task).Error
	//if err != nil {
	//	return nil, err
	//}
	//
	//taskConfig := s.config.GetTaskConfig()
	//idStr := strconv.FormatUint(uint64(task.ID), 10)
	//taskIDStr := strconv.FormatUint(uint64(task.TaskID), 10)
	//targetIDStr := strconv.FormatUint(uint64(task.TargetID), 10)
	//
	//discoveredHostFilePath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl.discovered")
	//crawlFilePath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl.scan")
	//scanFilePath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".scan")
	//
	//discoveredHostFile, err := os.Open(discoveredHostFilePath)
	//if err != nil {
	//	return nil, err
	//}
	//defer discoveredHostFile.Close()
	//discoveredHostLines := readLines(discoveredHostFile)
	//dh := c
	//
	//crawlFile, err := os.Open(crawlFilePath)
	//if err != nil {
	//	return nil, err
	//}
	//defer crawlFile.Close()
	//crawlLines := readLines(crawlFile)
	//cls := CrawlURL{
	//	TaskID:     task.TaskID,
	//	TargetID:   task.TargetID,
	//	ScanTarget: task.Target,
	//	CrawlURL:   crawlLines,
	//}
	//
	//scanFile, err := os.Open(scanFilePath)
	//if err != nil {
	//	return nil, err
	//}
	//defer scanFile.Close()
	//scanLines := readLines(scanFile)
	//
	//var newScanLines []map[string]interface{}
	//for _, line := range scanLines {
	//	var l map[string]interface{}
	//	err := json.Unmarshal([]byte(line), &l)
	//	if err != nil {
	//		s.logger.Errorf("json unmarshal error: %v", err)
	//		continue
	//	}
	//
	//	l["task_id"] = task.TaskID
	//	l["target_id"] = task.TargetID
	//	l["scan_target"] = task.Target
	//	l["target_vuln_id"] = uuid.New().String()
	//
	//	newScanLines = append(newScanLines, l)
	//}
	//
	//scanResults := make(map[string]interface{})
	//scanResults["discovered_host"] = dh
	//scanResults["crawl"] = cls
	//scanResults["scan"] = newScanLines
	//
	//return scanResults, err
	//// ------
	idStr := strconv.FormatUint(uint64(id), 10)
	taskIDStr := strconv.FormatUint(uint64(taskID), 10)
	targetIDStr := strconv.FormatUint(uint64(targetID), 10)

	crawledFilePath := path.Join(r.config.GetTaskConfig().DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl")

	crawledFile, err := os.Open(crawledFilePath)
	if err != nil {
		return nil, err
	}
	defer crawledFile.Close()

	var result crawlergo.SpiderResult
	err = json.NewDecoder(crawledFile).Decode(&result)

	scanResult := make(map[string]interface{})
	dh := make(map[string]interface{})
	dh["task_id"] = taskID
	dh["target_id"] = targetID
	dh["scan_target"] = target
	dh["discovered_host"] = result.AllDomainList
	scanResult["discovered_host"] = dh

	crawl := make(map[string]interface{})
	crawl["task_id"] = taskID
	crawl["target_id"] = targetID
	crawl["scan_target"] = target
	crawl["crawl_url"] = result.ReqList
	scanResult["crawl"] = crawl

	return scanResult, err
}

// GetScannedResult 获取扫描结果
func (r *Repository) GetScannedResult(id, taskID, targetID uint, target string) ([]map[string]interface{}, error) {
	idStr := strconv.FormatUint(uint64(id), 10)
	taskIDStr := strconv.FormatUint(uint64(taskID), 10)
	targetIDStr := strconv.FormatUint(uint64(targetID), 10)

	scannedFilePath := path.Join(r.config.GetTaskConfig().DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".scan")

	var result []map[string]interface{}

	scannedFile, err := os.Open(scannedFilePath)
	if err != nil {
		return result, err
	}
	defer scannedFile.Close()

	scanner := bufio.NewScanner(scannedFile)
	// 增加缓冲区大小
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024*5)
	for scanner.Scan() {
		line := scanner.Text()
		var tmp map[string]interface{}
		err := json.Unmarshal([]byte(line), &tmp)
		if err != nil {
			r.logger.Errorf("json unmarshal error: %v", err)
			continue
		}
		tmp["task_id"] = taskID
		tmp["target_id"] = targetID
		tmp["scan_target"] = target
		tmp["target_vuln_id"] = uuid.New().String()

		result = append(result, tmp)
	}
	if err := scanner.Err(); err != nil {
		return result, err
	}
	return result, err
}

// ExecuteNextTask 循环获取task表中的任务并执行
func (r *Repository) ExecuteNextTask(interval int) {
	// 定时器时长
	timer := time.NewTicker(time.Duration(interval) * time.Second)
	defer timer.Stop()

	ch := make(chan struct{}, 1)
	ch <- struct{}{}

	for {
		select {
		case <-timer.C:
			select {
			case <-ch:
				go func() {
					defer func() {
						ch <- struct{}{}
						timer.Reset(time.Duration(interval) * time.Second)
					}()
					// 获取下一个任务
					var nextTask entity.NodeTask
					err := r.mySqlite.Where("status = ?", "created").First(&nextTask).Error
					if err != nil {
						return
					}
					if nextTask.Target == "" {
						return
					}

					r.StartTask(&nextTask)
				}()
			}
		}
	}
}

func (r *Repository) UpdateSpiderResultFlag(id uint) {
	r.mySqlite.Model(&entity.NodeTask{}).Where("id = ?", id).Update("spider_result_pull_flag", true)
}
