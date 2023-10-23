package app

import (
	"webscan/config"
	"webscan/internal/domain/repository"
	"webscan/utils/log"
)

type MasterApp struct {
	config config.ConfigureInterface
	logger log.LoggerInterface
	mr     repository.MasterRepository
}

type MasterAppInterface interface {
	TaskDispatchLoop(interval int)
	CheckTaskStatusLoop(interval int)
	PullTaskResultLoop(interval int)
	CheckTaskIsOverLoop(interval int)
}

var _ MasterAppInterface = &MasterApp{}

func (m *MasterApp) TaskDispatchLoop(interval int) {
	m.mr.TaskDispatchLoop(interval)
}

func (m *MasterApp) CheckTaskStatusLoop(interval int) {
	m.mr.CheckTaskStatusLoop(interval)
}

func (m *MasterApp) PullTaskResultLoop(interval int) {
	m.mr.PullTaskResultLoop(interval)
}

func (m *MasterApp) CheckTaskIsOverLoop(interval int) {
	m.mr.CheckTaskIsOverLoop(interval)
}
