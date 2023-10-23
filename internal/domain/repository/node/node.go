package node

import "webscan/internal/domain/entity"

type RepositoryInterface interface {
	// NewTask 创建新任务
	NewTask(task *entity.NodeTask) (uint, error)
	// StartTask 开始任务
	StartTask(task *entity.NodeTask) error
	// StopTask 停止任务
	StopTask(task *entity.NodeTask) error
	DeleteTasks(tasks []entity.NodeTask)
	// PauseTask 暂停任务
	PauseTask(task *entity.NodeTask) error
	// ResumeTask 继续任务
	ResumeTask(task *entity.NodeTask) error

	// FindOneTask 查找任务
	FindOneTask(taskID, targetID uint) (entity.NodeTask, error)
	FindTask(taskID uint) ([]entity.NodeTask, error)
	// FindRunningTask 查找正在运行的任务
	FindRunningTask(taskID uint) ([]entity.NodeTask, error)
	// ResumePausedTask 继续暂停的任务
	ResumePausedTask(taskID uint) error
	// GetCrawledResult 获取任务的爬虫结果
	GetCrawledResult(id, taskID, targetID uint, target string) (map[string]interface{}, error)
	GetScannedResult(id, taskID, targetID uint, target string) ([]map[string]interface{}, error)

	ExecuteNextTask(interval int)

	UpdateSpiderResultFlag(id uint)
}
