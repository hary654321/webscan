package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"webscan/config"
	"webscan/internal/domain/entity"
	"webscan/internal/domain/repository"
	"webscan/internal/model"
	"webscan/utils/log"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
)

type MasterRepositoryImpl struct {
	config  config.ConfigureInterface
	logger  log.LoggerInterface
	myMySQL *gorm.DB
	elastic *elasticsearch.Client
}

var _ repository.MasterRepository = &MasterRepositoryImpl{}

func NewMasterRepositoryImpl(configure config.ConfigureInterface, logger log.LoggerInterface) *MasterRepositoryImpl {
	myMySQL, _ := InitDatabase(configure)
	elastic, _ := InitElastic(configure)

	return &MasterRepositoryImpl{
		config:  configure,
		logger:  logger,
		myMySQL: myMySQL,
		elastic: elastic,
	}
}

func (m *MasterRepositoryImpl) TaskDispatchLoop(interval int) {
	timer := time.NewTimer(time.Duration(interval) * time.Second)
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
						timer.Reset(5 * time.Second)
					}()
					// 1. 遍历task表
					var tasks []model.Task
					err := m.myMySQL.Find(&tasks).Error
					if err != nil {
						m.logger.Errorf("find all task error: %s", err.Error())
						return
					}

					for _, task := range tasks {
						var taskTargets []model.TaskTarget
						err = m.myMySQL.Where("task_id = ?", task.ID).Find(&taskTargets).Error
						if err != nil {
							m.logger.Errorf("find task target error: %s", err.Error())
							continue
						}
						//targets := task.TaskTargets
						for _, target := range taskTargets {
							_, err := m.FindScanRecordByTaskAndTargetID(task.ID, target.ID)

							if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
								m.logger.Errorf("find scan by target id error: %s", err.Error())
								continue
							}
							if errors.Is(err, gorm.ErrRecordNotFound) {
								//  插入一条新的scan记录
								newScan := model.Scan{
									TaskID:     task.ID,
									Name:       task.Name,
									Node:       task.Node,
									TargetID:   target.ID,
									ScanTarget: target.Target,
									Status:     0,
									Pull:       0,
									StartTime:  time.Now(),
									EndTime:    time.Now(),
								}
								err = m.InsertScanRecord(newScan)
								if err != nil {
									m.logger.Errorf("insert scan: %v error: %s", newScan, err.Error())
									continue
								}

								// 将任务分配给节点
								nodeTask, err := m.buildNodeTask(task, target)
								if err != nil {
									m.logger.Errorf("build node task error: %s", err.Error())
									continue
								}

								//var taskExchange entity.NodeTask
								//taskExchange.TaskID = task.ID
								//taskExchange.TargetID = target.ID
								//taskExchange.Target = target.Target
								resp, err := resty.New().R().
									SetHeader("Content-Type", "application/json").
									SetBody(nodeTask).
									Post(fmt.Sprintf("http://%s:18080/api/node/v1/new_task", task.Node))

								if err != nil {
									m.logger.Errorf("post new task to node error: %s", err.Error())
									continue
								}
								if task.Status == 0 {
									err = m.myMySQL.Model(&model.Task{}).Where("id = ?", task.ID).Update("status", 1).Error
									if err != nil {
										m.logger.Errorf("update task status error: %s", err.Error())
									}
								}
								err = m.myMySQL.Model(&model.Scan{}).Where("task_id = ? and target_id = ?", task.ID, target.ID).Update("status", 1).Error
								if err != nil {
									m.logger.Errorf("update scan status error: %s", err.Error())
								}
								m.logger.Infof("resp: %v", resp)
							}
						}
					}
				}()
			}
		}
	}
}

// CheckTaskStatusLoop 检查任务状态
func (m *MasterRepositoryImpl) CheckTaskStatusLoop(interval int) {
	timer := time.NewTimer(time.Duration(interval) * time.Second)
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
					scans, err := m.FindRunningScan()
					if err != nil {
						m.logger.Errorf("find running scan error: %s", err.Error())
						return
					}
					for _, scan := range scans {
						body := fmt.Sprintf(`{"task_id": %d, "target_id": %d}`, scan.TaskID, scan.TargetID)
						resp, err := resty.New().R().
							SetHeader("Content-Type", "application/json").
							SetBody(body).
							Post(fmt.Sprintf("http://%s:18080/api/node/v1/task_state", scan.Node))
						if err != nil {
							m.logger.Errorf("get task status error: %s, body: %v", err.Error(), resp.Body())
							continue
						}
						var respBody entity.Response
						err = json.Unmarshal(resp.Body(), &respBody)
						if err != nil {
							m.logger.Errorf("unmarshal response body error: %s", err.Error())
							continue
						}
						m.logger.Infof("===> resp body: %v", respBody)
						if respBody.Success && respBody.Data.(string) == "finished" {

							err = m.myMySQL.Model(&model.Scan{}).Where("task_id = ? and target_id = ?", scan.TaskID, scan.TargetID).Update("status", 2).Error
							if err != nil {
								m.logger.Errorf("update scan status error: %s", err.Error())
							}
						}
					}
				}()
			}
		}
	}
}

func (m *MasterRepositoryImpl) PullTaskResultLoop(interval int) {
	timer := time.NewTimer(time.Duration(interval) * time.Second)
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

					scans, err := m.FindNeedToPullScanResult()
					if err != nil {
						m.logger.Errorf("find need to pull scan result error: %s", err.Error())
						return
					}
					for _, scan := range scans {
						body := fmt.Sprintf(`{"task_id": %d, "target_id": %d}`, scan.TaskID, scan.TargetID)
						resp, err := resty.New().R().
							SetHeader("Content-Type", "application/json").
							SetBody(body).
							Post(fmt.Sprintf("http://%s:18080/api/node/v1/pull_result", scan.Node))

						if err != nil {
							m.logger.Errorf("pull task result error: %s", err.Error())
							continue
						}
						m.logger.Infof("------------> resp: %s", resp.Body())
						// 保存到ES中
						// 准备请求
						var respBody map[string]interface{}
						err = json.Unmarshal(resp.Body(), &respBody)
						if err != nil {
							m.logger.Errorf("unmarshal response body error: %s", err.Error())
							continue
						}
						if !respBody["success"].(bool) {
							m.logger.Errorf("pull task result error: %s", respBody["error"].(string))
							continue
						}
						/*
							scanResults["discovered_host"] = dh
							scanResults["crawl"] = cls
							scanResults["scan"] = newScanLines
						*/
						//var respData map[string]interface{}
						//err = json.Unmarshal([]byte(respBody["data"].(string)), &respData)
						//if err != nil {
						//	m.logger.Errorf("unmarshal respData error: %s", err.Error())
						//	continue
						//}

						respData := respBody["data"].(map[string]interface{})

						{
							discoveredHostJson, err := json.Marshal(respData["discovered_host"])
							if err != nil {
								fmt.Printf("Error marshaling JSON: %v\n", err)
								return
							}
							req := esapi.IndexRequest{
								Index:   "discovered_host_index",
								Body:    strings.NewReader(string(discoveredHostJson)),
								Refresh: "true", // 立即刷新
							}

							// 执行请求
							res, err := req.Do(context.Background(), m.elastic)
							if err != nil {
								m.logger.Errorf("Error indexing document: %s", err)
							} else {
								defer res.Body.Close()
							}

							if res.IsError() {
								m.logger.Errorf("Error indexing document: %s", res.Status())
							} else {
								m.logger.Infof("index document success")
							}
						}

						{
							crawlJson, err := json.Marshal(respData["crawl"])
							if err != nil {
								m.logger.Errorf("marshal crawl error: %s", err.Error())
								continue
							}
							req := esapi.IndexRequest{
								Index:   "craw_index",
								Body:    strings.NewReader(string(crawlJson)),
								Refresh: "true", // 立即刷新
							}

							// 执行请求
							res, err := req.Do(context.Background(), m.elastic)
							if err != nil {
								m.logger.Errorf("Error indexing document: %s", err)
							} else {
								defer res.Body.Close()
							}
							if res.IsError() {
								m.logger.Errorf("Error indexing document: %s", res.Status())
							} else {
								m.logger.Infof("index document success")
							}
						}

						{
							if respData["scan"] != nil {
								scanLines := respData["scan"].([]interface{})
								for _, scanLine := range scanLines {
									slb, err := json.Marshal(scanLine)
									if err != nil {
										m.logger.Errorf("marshal crawl error: %s", err.Error())
										continue
									}
									req := esapi.IndexRequest{
										Index:   "scan_index",
										Body:    strings.NewReader(string(slb)),
										Refresh: "true", // 立即刷新
									}

									// 执行请求
									res, err := req.Do(context.Background(), m.elastic)
									if err != nil {
										m.logger.Errorf("Error indexing document: %s", err)
									} else {
										defer res.Body.Close()
									}

									if res.IsError() {
										m.logger.Errorf("Error indexing document: %s", res.Status())
									} else {
										m.logger.Infof("index document success")
									}
								}
							}
						}
						err = m.myMySQL.Model(&model.Scan{}).Where("task_id = ? and target_id = ?", scan.TaskID, scan.TargetID).Update("pull", 1).Error
						if err != nil {
							m.logger.Errorf("update scan pull status error: %s", err.Error())
							continue
						}
						m.logger.Infof("resp: %s", string(resp.Body()))
					}
				}()
			}
		}
	}
}

func (m *MasterRepositoryImpl) CheckTaskIsOverLoop(interval int) {
	timer := time.NewTimer(time.Duration(interval) * time.Second)
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

					var tasks []model.Task
					err := m.myMySQL.Where("status", 1).Find(&tasks).Error
					if err != nil {
						m.logger.Errorf("find all task error: %s", err.Error())
						return
					}
					for _, task := range tasks {
						var taskTarget []model.TaskTarget
						err = m.myMySQL.Where("task_id = ?", task.ID).Find(&taskTarget).Error
						if err != nil {
							m.logger.Errorf("find task target error: %s", err.Error())
							continue
						}
						if len(taskTarget) == 0 {
							m.logger.Errorf("task %d has no target", task.ID)
							continue
						}

						taskOver := true
						for _, target := range taskTarget {
							var scan model.Scan
							err = m.myMySQL.Where("task_id = ? and target_id = ?", task.ID, target.ID).
								Where("status = ? and pull = ?", 2, 1).
								First(&scan).Error

							if err != nil {
								taskOver = false
								// m.logger.Errorf("find scan record error: %s", err.Error())
								break
							}
							if scan.Pull == 0 {
								taskOver = false
								// 如果还有任务未 pull 结果
								break
							}
						}

						if taskOver {
							m.logger.Infof("================> task %d is over", task.ID)
							err = m.myMySQL.Model(&model.Task{}).Where("id = ?", task.ID).Update("status", 2).Error
							if err != nil {
								m.logger.Errorf("update Task status error: %s", err.Error())
								continue
							}
						}
					}
				}()
			}

		}
	}
}

func (m *MasterRepositoryImpl) FindScanRecordByTaskAndTargetID(taskID, targetID uint) (model.Scan, error) {
	var scan model.Scan
	err := m.myMySQL.Where("task_id = ? AND target_id = ?", taskID, targetID).First(&model.Scan{}).Error
	return scan, err
}

func (m *MasterRepositoryImpl) InsertScanRecord(scan model.Scan) error {
	err := m.myMySQL.Create(&scan).Error
	return err
}

func (m *MasterRepositoryImpl) FindNeedToPullScanResult() ([]model.Scan, error) {
	var scans []model.Scan
	err := m.myMySQL.Where("status = ? and pull = ?", 2, 0).Find(&scans).Error
	return scans, err
}

func (m *MasterRepositoryImpl) FindRunningScan() ([]model.Scan, error) {
	var scans []model.Scan
	err := m.myMySQL.Where("status = ? ", 1).Find(&scans).Error
	return scans, err
}

func (m *MasterRepositoryImpl) buildNodeTask(task model.Task, target model.TaskTarget) (entity.NodeNewTaskArgs, error) {
	var nodeTask entity.NodeNewTaskArgs
	nodeTask.TaskID = task.ID
	nodeTask.TargetID = target.ID
	nodeTask.Target = target.Target

	var spiderConfig entity.SpiderConfigs
	err := json.Unmarshal([]byte(task.SpiderConfig), &spiderConfig)
	if err != nil {
		return nodeTask, err
	}
	nodeTask.SpiderConfig = spiderConfig

	var scannerConfig entity.ScannerConfigs
	err = json.Unmarshal([]byte(task.ScannerConfig), &scannerConfig)
	if err != nil {
		return nodeTask, err
	}
	nodeTask.ScannerConfig = scannerConfig
	return nodeTask, nil
}
