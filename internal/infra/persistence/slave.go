package persistence

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"gorm.io/gorm"
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"
	"webscan/config"
	"webscan/internal/domain/entity"
	"webscan/internal/domain/repository"
	"webscan/pkg/katana"
	"webscan/pkg/nuclei"
	"webscan/utils/cache"
	"webscan/utils/log"
)

type SlaveRepositoryImpl struct {
	config   config.ConfigureInterface
	logger   log.LoggerInterface
	katana   katana.KatanaInterface
	nuclei   nuclei.ScannerInterface
	mySqlite *gorm.DB

	nodeStatus cache.NodeStatusInterface
}

var _ repository.SlaveRepository = &SlaveRepositoryImpl{}

func NewSlavePersistence(logger log.LoggerInterface, configure config.ConfigureInterface, katanaInterface katana.KatanaInterface, nucleiInterface nuclei.ScannerInterface, nodeStatus cache.NodeStatusInterface) *SlaveRepositoryImpl {
	_, mySqlite := InitDatabase(configure)
	mySqlite.AutoMigrate(&entity.SlaveTask{})
	return &SlaveRepositoryImpl{
		config:     configure,
		logger:     logger,
		katana:     katanaInterface,
		nuclei:     nucleiInterface,
		mySqlite:   mySqlite,
		nodeStatus: nodeStatus,
	}
}

func (s *SlaveRepositoryImpl) NewTask(task *entity.SlaveTask) (uint, error) {
	tx := s.mySqlite.Create(task)
	return task.ID, tx.Error
}

// GetTaskByTaskAndTargetID 使用taskID查询多条记录
func (s *SlaveRepositoryImpl) GetTaskByTaskAndTargetID(taskID, targetID uint) (entity.SlaveTask, error) {
	var task entity.SlaveTask
	err := s.mySqlite.Where("task_id = ? and target_id = ?", taskID, targetID).First(&task).Error
	return task, err
}

func (s *SlaveRepositoryImpl) UpdateTaskStatus(id uint, key, value string) error {
	updateFields := map[string]interface{}{
		key: value,
	}
	err := s.mySqlite.Model(&entity.SlaveTask{}).Where("id = ?", id).Updates(updateFields).Error
	return err
}

// StartKatana 启动Katana爬取目标
func (s *SlaveRepositoryImpl) StartKatana(task *entity.SlaveTask) error {
	taskConfig := s.config.GetTaskConfig()

	// dir/1-1-1.crawl
	idStr := strconv.FormatUint(uint64(task.ID), 10)
	taskIDStr := strconv.FormatUint(uint64(task.TaskID), 10)
	targetIDStr := strconv.FormatUint(uint64(task.TargetID), 10)
	katanaOutputPath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl")

	err := s.katana.Start(task, katanaOutputPath)
	if err != nil {
		return err
	}

	// 读取输出文件，解析url
	// 打开文件
	file, err := os.Open(katanaOutputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 按行读取文件内容
	scanner := bufio.NewScanner(file)
	//targetDirPath := path.Dir(task.TargetPath)
	targetParse, _ := url.Parse(task.Target)

	scanURL := make(map[string]struct{})
	discoveredHost := make(map[string]struct{})

	// 逐行读取文件内容
	for scanner.Scan() {
		line := scanner.Text()
		u, parseErr := url.Parse(line)
		if parseErr != nil {
			continue
		}
		newURL := url.URL{
			Scheme: u.Scheme,
			Opaque: u.Opaque,
			User:   u.User,
			Host:   u.Host,
		}

		if u.Host == targetParse.Host {
			tmpPath := u.Path
			if tmpPath == "/" {
				tmpPath = ""
			}
			cs := path.Dir(tmpPath)
			if cs == "." {
				cs = ""
			}
			if cs == task.CrawlScope {
				s.logger.Infof("******* %s", line)
				scanURL[line] = struct{}{}
			}

		}
		//if newURL.String() == task.TargetHost && path.Dir(u.Path) == targetDirPath {
		//	scanURL[line] = struct{}{}
		//}
		if newURL.String() != task.TargetHost && newURL.String() != "" {
			discoveredHost[newURL.String()] = struct{}{}
		}
	}
	// 检查是否出现了扫描错误
	if err := scanner.Err(); err != nil {
		return err
	}

	// scan file path: dir/1-1.crawl.discovered
	discoveredHostFilePath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl.discovered")
	discoveredHostFile, err := os.Create(discoveredHostFilePath)
	if err != nil {
		return err
	}
	defer discoveredHostFile.Close()

	for k, _ := range discoveredHost {
		discoveredHostFile.WriteString(k + "\n")
	}

	// scan file path: dir/1-1.crawl.scan
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

type DiscoveredHost struct {
	TaskID         uint     `json:"task_id"`
	TargetID       uint     `json:"target_id"`
	ScanTarget     string   `json:"scan_target"`
	DiscoveredHost []string `json:"discovered_host"`
}

type CrawlURL struct {
	TaskID     uint     `json:"task_id"`
	TargetID   uint     `json:"target_id"`
	ScanTarget string   `json:"scan_target"`
	CrawlURL   []string `json:"craw_url"`
}

// StartNuclei 启动Nuclei扫描目标
func (s *SlaveRepositoryImpl) StartNuclei(task *entity.SlaveTask) error {
	idStr := strconv.FormatUint(uint64(task.ID), 10)
	taskIDStr := strconv.FormatUint(uint64(task.TaskID), 10)
	targetIDStr := strconv.FormatUint(uint64(task.TargetID), 10)

	scanFilePath := path.Join(s.config.GetTaskConfig().DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl.scan")
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

	taskConfig := s.config.GetTaskConfig()
	nucleiOutputPath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".scan")

	//获取模板目录
	templatesDir := s.config.GetAppConfig().Templates
	s.logger.Infof("[StartNuclei] task id: %d, target id: %d, templatesDir: %s", task.TaskID, task.TargetID, templatesDir)
	return s.nuclei.Start(task, templatesDir, []string{scans[0]}, nucleiOutputPath)
}

// StartTask 启动任务：启动Katana爬取目标，启动Nuclei扫描目标
func (s *SlaveRepositoryImpl) StartTask(taskID, targetID uint) {
	targetTask, err := s.GetTaskByTaskAndTargetID(taskID, targetID)
	if err != nil {
		s.logger.Errorf("[StartTask] task id: %d, target id: %d, get task error: %v", taskID, targetID, err)
		return
	}

	err = s.StartKatana(&targetTask)
	if err != nil {
		s.logger.Errorf("[StartTask] task id: %d, target id: %d, start katana error: %v", taskID, targetID, err)
		return
	}
	err = s.StartNuclei(&targetTask)
	if err != nil {
		s.logger.Errorf("[StartTask] task id: %d, target id: %d, start nuclei error: %v", taskID, targetID, err)
		return
	}
	s.UpdateTaskStatus(targetTask.ID, "status", "finished")
}

// GetNextTask 获取下一个要执行的任务
func (s *SlaveRepositoryImpl) GetNextTask() (entity.SlaveTask, error) {
	var task entity.SlaveTask
	err := s.mySqlite.Where("status != ?", "finished").First(&task).Error
	return task, err
}

func (s *SlaveRepositoryImpl) StartTaskLoop() {
	timer := time.NewTicker(10 * time.Second)
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
						timer.Reset(10 * time.Second)
					}()

					task, err := s.GetNextTask()
					if err != nil {
						return
					}
					if task.Target == "" {
						return
					}
					if task.Status == "finished" {
						return
					}

					ns := cache.NodeStatus{
						Status:   "running",
						TaskID:   task.TaskID,
						TargetID: task.TargetID,
					}
					s.nodeStatus.Set("self", ns)
					s.StartTask(task.TaskID, task.TargetID)
				}()
			}
		}
	}
}

func (s *SlaveRepositoryImpl) GetNodeStatus() *cache.NodeStatus {
	return s.nodeStatus.Get("self")
}

func (s *SlaveRepositoryImpl) UpdateNodeSystemStatusLoop() {
	// 获取 CPU 使用率
	cpuPercent, err := cpu.Percent(time.Second, false)
	cpuUsage := ""
	if err == nil {
		cpuUsage = fmt.Sprintf("%.2f%%", cpuPercent[0])
	}

	// 获取内存使用率
	memInfo, err := mem.VirtualMemory()
	memInfoTotal := ""
	memInfoFree := ""
	memInfoUsed := ""
	memInfoUsedPercent := ""
	if err == nil {
		memInfoTotal = fmt.Sprintf("%v", memInfo.Total)
		memInfoFree = fmt.Sprintf("%v", memInfo.Free)
		memInfoUsed = fmt.Sprintf("%v", memInfo.Used)
		memInfoUsedPercent = fmt.Sprintf("%.2f%%", memInfo.UsedPercent)
	}

	//// 获取硬盘使用率
	//partitions, err := disk.Partitions(true)
	//var diskUsage []map[string]string
	//if err == nil {
	//	for _, partition := range partitions {
	//		usage, err := disk.Usage(partition.Mountpoint)
	//		if err == nil {
	//			diskUsage = append(diskUsage, map[string]string{
	//				"mountpoint":   partition.Mountpoint,
	//				"total":        fmt.Sprintf("%v", usage.Total),
	//				"free":         fmt.Sprintf("%v", usage.Free),
	//				"used":         fmt.Sprintf("%v", usage.Used),
	//				"used_percent": fmt.Sprintf("%.2f%%", usage.UsedPercent),
	//			})
	//		}
	//	}
	//}
	s.nodeStatus.SetSystemStatus("self", cpuUsage, memInfoTotal, memInfoFree, memInfoUsed, memInfoUsedPercent)
}

func (s *SlaveRepositoryImpl) GetTaskScanResult(taskID, targetID uint) (map[string]interface{}, error) {
	var task entity.SlaveTask
	err := s.mySqlite.Where("task_id = ? and target_id = ?", taskID, targetID).Find(&task).Error
	if err != nil {
		return nil, err
	}

	taskConfig := s.config.GetTaskConfig()
	idStr := strconv.FormatUint(uint64(task.ID), 10)
	taskIDStr := strconv.FormatUint(uint64(task.TaskID), 10)
	targetIDStr := strconv.FormatUint(uint64(task.TargetID), 10)

	discoveredHostFilePath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl.discovered")
	crawlFilePath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".crawl.scan")
	scanFilePath := path.Join(taskConfig.DataDir, idStr+"-"+taskIDStr+"-"+targetIDStr+".scan")

	discoveredHostFile, err := os.Open(discoveredHostFilePath)
	if err != nil {
		return nil, err
	}
	defer discoveredHostFile.Close()
	discoveredHostLines := readLines(discoveredHostFile)
	dh := DiscoveredHost{
		TaskID:         task.TaskID,
		TargetID:       task.TargetID,
		ScanTarget:     task.Target,
		DiscoveredHost: discoveredHostLines,
	}

	crawlFile, err := os.Open(crawlFilePath)
	if err != nil {
		return nil, err
	}
	defer crawlFile.Close()
	crawlLines := readLines(crawlFile)
	cls := CrawlURL{
		TaskID:     task.TaskID,
		TargetID:   task.TargetID,
		ScanTarget: task.Target,
		CrawlURL:   crawlLines,
	}

	scanFile, err := os.Open(scanFilePath)
	if err != nil {
		return nil, err
	}
	defer scanFile.Close()
	scanLines := readLines(scanFile)

	var newScanLines []map[string]interface{}
	for _, line := range scanLines {
		var l map[string]interface{}
		err := json.Unmarshal([]byte(line), &l)
		if err != nil {
			s.logger.Errorf("json unmarshal error: %v", err)
			continue
		}

		l["task_id"] = task.TaskID
		l["target_id"] = task.TargetID
		l["scan_target"] = task.Target
		l["target_vuln_id"] = uuid.New().String()

		newScanLines = append(newScanLines, l)
	}

	scanResults := make(map[string]interface{})
	scanResults["discovered_host"] = dh
	scanResults["crawl"] = cls
	scanResults["scan"] = newScanLines

	return scanResults, err
}

func (s *SlaveRepositoryImpl) GetTaskScanStatus(taskID, targetID uint) (string, error) {
	var task entity.SlaveTask
	err := s.mySqlite.Where("task_id = ? and target_id = ?", taskID, targetID).Find(&task).Error
	return task.Status, err
}

func readLines(f io.Reader) []string {
	scanner := bufio.NewScanner(f)

	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	//// 检查扫描过程中是否出现错误
	//if err := scanner.Err(); err != nil {
	//
	//}
	return lines
}
