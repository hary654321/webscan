package repository

import (
	"webscan/internal/domain/entity"
	"webscan/utils/cache"
)

type SlaveRepository interface {
	// NewTask 创建新任务
	NewTask(task *entity.SlaveTask) (uint, error)
	// GetTaskByTaskAndTargetID 获取一个任务
	GetTaskByTaskAndTargetID(taskID, targetID uint) (entity.SlaveTask, error)
	// UpdateTaskStatus 更新任务状态
	UpdateTaskStatus(id uint, key, value string) error
	// StartKatana 启动Katana爬取目标
	StartKatana(task *entity.SlaveTask) error
	// StartNuclei 启动Nuclei扫描目标
	StartNuclei(task *entity.SlaveTask) error
	// StartTask 启动任务：启动Katana爬取目标，启动Nuclei扫描目标
	StartTask(taskID, targetID uint)

	StartTaskLoop()
	GetNodeStatus() *cache.NodeStatus
	UpdateNodeSystemStatusLoop()

	GetNextTask() (entity.SlaveTask, error)

	GetTaskScanResult(taskID, targetID uint) (map[string]interface{}, error)
	GetTaskScanStatus(taskID, targetID uint) (string, error)
}
