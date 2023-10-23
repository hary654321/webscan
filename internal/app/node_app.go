package app

import (
	"github.com/pkg/errors"
	"webscan/config"
	"webscan/internal/domain/entity"
	"webscan/internal/domain/repository/node"
	"webscan/pkg/crawlergo"
	myNuclei "webscan/pkg/nuclei"
	"webscan/utils/log"
)

type NodeAppInterface interface {
	// NewTask 创建新任务
	NewTask(task *entity.NodeTask) (uint, error)
	// StartTask 开始任务
	StartTask(task *entity.NodeTask) error
	// StopTask 停止任务
	StopTask(task *entity.NodeTask) error
	DeleteTasks(tasks []entity.NodeTask)
	// PauseTask 暂停任务
	PauseTask(taskID uint) error
	// ResumeTask 继续任务
	ResumeTask(taskID uint) error
	// PullResult 获取任务结果
	PullResult(taskID, targetID uint) (map[string]interface{}, error)

	// FindOneTask 查找任务
	FindOneTask(taskID, targetID uint) (entity.NodeTask, error)

	FindTask(taskID uint) ([]entity.NodeTask, error)
}

type NodeApp struct {
	config       config.ConfigureInterface
	logger       log.LoggerInterface
	nodeRepoFace node.RepositoryInterface
	crawlergo    crawlergo.SpiderInterface
	nuclei       myNuclei.ScannerInterface
}

func NewNodeApp(
	configFace config.ConfigureInterface,
	logFace log.LoggerInterface,
	nodeRepoFace node.RepositoryInterface,
	crawlergo crawlergo.SpiderInterface,
	nuclei myNuclei.ScannerInterface,
) *NodeApp {
	return &NodeApp{
		config:       configFace,
		logger:       logFace,
		nodeRepoFace: nodeRepoFace,
		crawlergo:    crawlergo,
		nuclei:       nuclei,
	}
}

func (app *NodeApp) NewTask(task *entity.NodeTask) (uint, error) {
	// 插入一个新任务
	return app.nodeRepoFace.NewTask(task)
}

func (app *NodeApp) StartTask(task *entity.NodeTask) error {
	return app.nodeRepoFace.StartTask(task)
}

func (app *NodeApp) StopTask(task *entity.NodeTask) error {
	return app.nodeRepoFace.StopTask(task)
}

func (app *NodeApp) DeleteTasks(tasks []entity.NodeTask) {
	app.nodeRepoFace.DeleteTasks(tasks)
}

// PauseTask 暂停任务
func (app *NodeApp) PauseTask(taskID uint) error {
	runningTask, err := app.nodeRepoFace.FindRunningTask(taskID)
	if err != nil {
		return err
	}

	for _, rt := range runningTask {
		err = app.nodeRepoFace.PauseTask(&rt)
	}
	return nil
}

func (app *NodeApp) ResumeTask(taskID uint) error {
	return app.nodeRepoFace.ResumePausedTask(taskID)
}

func (app *NodeApp) PullResult(taskID, targetID uint) (map[string]interface{}, error) {
	task, err := app.nodeRepoFace.FindOneTask(taskID, targetID)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{}

	//爬虫执行完成，读取爬虫的结果
	if task.Status == "crawled" && !task.SpiderResultPullFlag {
		spiderResult, err := app.nodeRepoFace.GetCrawledResult(task.ID, task.TaskID, task.TargetID, task.Target)
		if err != nil {
			return result, err
		}
		for k, v := range spiderResult {
			result[k] = v
		}
		result["scan"] = []map[string]interface{}{}
		app.nodeRepoFace.UpdateSpiderResultFlag(task.ID)
		return result, nil
	}

	// 1.获取内存状态，如果爬虫执行完成，返回爬虫的结果，如果全部执行完成返回全部的结果
	if task.Status == "finished" {
		//全部执行完成读取所有文件并生成一个JSON返回
		spiderResult, err := app.nodeRepoFace.GetCrawledResult(task.ID, task.TaskID, task.TargetID, task.Target)
		if err != nil {
			return result, err
		}
		for k, v := range spiderResult {
			result[k] = v
		}
		scannerResult, err := app.nodeRepoFace.GetScannedResult(task.ID, task.TaskID, task.TargetID, task.Target)
		if err != nil {
			return result, err
		}
		result["scan"] = scannerResult
		return result, nil
	}
	return nil, errors.New("task_not_finished")
}

func (app *NodeApp) FindOneTask(taskID, targetID uint) (entity.NodeTask, error) {
	return app.nodeRepoFace.FindOneTask(taskID, targetID)
}

func (app *NodeApp) FindTask(taskID uint) ([]entity.NodeTask, error) {
	return app.nodeRepoFace.FindTask(taskID)
}
