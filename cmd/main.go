package main

import (
	"webscan/config"
	"webscan/internal/app"
	cache2 "webscan/internal/infra/cache"
	nodeInfra "webscan/internal/infra/node"
	"webscan/internal/infra/persistence"
	masterInterface "webscan/internal/interfaces/master"
	nodeInterface "webscan/internal/interfaces/node"
	"webscan/pkg/crawlergo"
	"webscan/pkg/nuclei"
	"webscan/utils/log"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
)

func main() {
	// init config
	conf := config.NewConfig("config.yaml")
	conf.InitCheck()

	// init logger
	logger := log.NewLogger(conf)
	logger.Info("logger init success!")

	// katana spider
	//katanaInterface := katana.NewKatana()
	nucleiInterface := nuclei.NewNuclei()

	stataCacheInterface := cache2.NewStateCache()
	// crawlergo spider
	spiderInterface := crawlergo.NewSpider()
	nodePersistence := nodeInfra.NewNodePersistence(conf, logger, spiderInterface, nucleiInterface, stataCacheInterface)
	go nodePersistence.ExecuteNextTask(10)

	// 如果当前是 master 节点，需要初始化 master 数据库连接，启动 master 服务
	if conf.App.Role == "master" {
		//masterNodeCache := cache.NewNodeStatusCache()
		masterPersistence := persistence.NewMasterRepositoryImpl(conf, logger)
		go masterPersistence.TaskDispatchLoop(10)
		// 检查scan表中的任务是否结束
		go masterPersistence.CheckTaskStatusLoop(10)
		go masterPersistence.PullTaskResultLoop(10)
		go masterPersistence.CheckTaskIsOverLoop(10)
	}

	nodeApp := app.NewNodeApp(conf, logger, nodePersistence, spiderInterface, nucleiInterface)
	nodeAPI := nodeInterface.NewNode(conf, logger, nodeApp)
	r := gin.Default()
	ginpprof.Wrap(r)
	nodeGroup := r.Group("/api/node/v1")
	{
		//创建任务
		nodeGroup.POST("/new_task", nodeAPI.NewTaskHandler)
		//拉取结果
		nodeGroup.POST("/pull_result", nodeAPI.PullResultHandler)
		//暂停任务
		nodeGroup.POST("/pause_task", nodeAPI.PauseTaskHandler)
		//恢复执行
		nodeGroup.POST("/resume_task", nodeAPI.ResumeTaskHandler)
		////停止任务
		//nodeGroup.POST("/stop_task", nodeAPI.StopTaskHandler)
		//删除任务
		nodeGroup.POST("/delete_task", nodeAPI.DeleteTaskHandler)
		//查询任务状态
		nodeGroup.POST("/task_state", nodeAPI.TaskStateHandler)
	}

	if conf.App.Role == "master" {
		masterAPI := masterInterface.NewMaster(conf, logger, nodeApp)
		masterGroup := r.Group("/api/master/v1")
		{
			//暂停任务
			masterGroup.POST("/pause_task", masterAPI.PauseTaskHandler)
			//恢复执行
			masterGroup.POST("/resume_task", masterAPI.ResumeTaskHandler)
			//删除任务
			masterGroup.POST("/delete_task", masterAPI.DeleteTaskHandler)
		}
	}
	r.Run(":" + conf.App.Port)
}
