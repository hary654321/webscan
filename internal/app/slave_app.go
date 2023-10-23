package app

import (
	"webscan/config"
	"webscan/internal/domain/entity"
	"webscan/internal/domain/repository"
	"webscan/utils/cache"
	"webscan/utils/log"
)

type SlaveApp struct {
	config config.ConfigureInterface
	logger log.LoggerInterface
	sr     repository.SlaveRepository
}

type SlaveAppInterface interface {
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

var _ SlaveAppInterface = &SlaveApp{}

func (s *SlaveApp) NewTask(task *entity.SlaveTask) (uint, error) {
	return s.sr.NewTask(task)
}

func (s *SlaveApp) GetTaskByTaskAndTargetID(taskID, targetID uint) (entity.SlaveTask, error) {
	return s.sr.GetTaskByTaskAndTargetID(taskID, targetID)
}

func (s *SlaveApp) UpdateTaskStatus(id uint, key, value string) error {
	return s.sr.UpdateTaskStatus(id, key, value)
}

func (s *SlaveApp) StartKatana(task *entity.SlaveTask) error {
	return s.sr.StartKatana(task)
}

func (s *SlaveApp) StartNuclei(task *entity.SlaveTask) error {
	return s.sr.StartNuclei(task)
}

func (s *SlaveApp) StartTask(taskID, targetID uint) {
	s.sr.StartTask(taskID, targetID)
}

func (s *SlaveApp) StartTaskLoop() {
	s.sr.StartTaskLoop()
}

func (s *SlaveApp) GetNodeStatus() *cache.NodeStatus {
	return s.sr.GetNodeStatus()
}

func (s *SlaveApp) UpdateNodeSystemStatusLoop() {
	s.sr.UpdateNodeSystemStatusLoop()
}

func (s *SlaveApp) GetNextTask() (entity.SlaveTask, error) {
	return s.sr.GetNextTask()
}

func (s *SlaveApp) GetTaskScanResult(taskID, targetID uint) (map[string]interface{}, error) {
	return s.sr.GetTaskScanResult(taskID, targetID)
}

func (s *SlaveApp) GetTaskScanStatus(taskID, targetID uint) (string, error) {
	return s.sr.GetTaskScanStatus(taskID, targetID)
}
