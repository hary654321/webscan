package node

import (
	"webscan/config"
	"webscan/internal/app"
	"webscan/utils/log"
)

type Node struct {
	config config.ConfigureInterface
	logger log.LoggerInterface
	app    app.NodeAppInterface
}

func NewNode(
	config config.ConfigureInterface,
	logger log.LoggerInterface,
	app app.NodeAppInterface,
) *Node {
	return &Node{
		config: config,
		logger: logger,
		app:    app,
	}
}
