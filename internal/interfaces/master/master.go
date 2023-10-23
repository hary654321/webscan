package master

import (
	"webscan/config"
	"webscan/internal/app"
	"webscan/utils/log"
)

type Master struct {
	config config.ConfigureInterface
	logger log.LoggerInterface

	app app.NodeAppInterface
}

func NewMaster(config config.ConfigureInterface, logger log.LoggerInterface, app app.NodeAppInterface) *Master {
	return &Master{
		config: config,
		logger: logger,
		app:    app,
	}
}

type RequestArgs struct {
	TaskID uint   `json:"task_id" required:"true" comment:"任务ID"`
	NodeIP string `json:"node_ip" required:"true" comment:"节点IP"`
}
