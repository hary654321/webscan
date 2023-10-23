package v1

import (
	"webscan/config"
	"webscan/internal/app"
	"webscan/utils/log"
)

//type SlaveInterface interface {
//	SlaveNewTaskHandler(c *gin.Context)
//}

type Slave struct {
	config config.ConfigureInterface
	logger log.LoggerInterface

	app app.SlaveAppInterface
}

func NewSlave(config config.ConfigureInterface, logger log.LoggerInterface, app app.SlaveAppInterface) *Slave {
	return &Slave{
		config: config,
		logger: logger,
		app:    app,
	}
}
