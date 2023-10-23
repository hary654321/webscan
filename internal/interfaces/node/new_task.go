package node

import (
	"encoding/json"
	"net/url"
	"path"
	"time"
	"webscan/internal/domain/entity"

	"github.com/gin-gonic/gin"
)

// NewTaskHandler
// @Summary 创建扫描任务接口
// @Description 扫描任务分为两个阶段：1.爬虫，2.扫描。接口接收到的参数应该包含三部份：1.任务参数，2.爬虫参数，3.扫描参数。
// @Tags 创建扫描任务
// @Accept json
// @Produce json
// @Param args body TaskExchange true "args"
// @Success 200 {object} Response
// @Router /api/node/new_task [post]
func (n *Node) NewTaskHandler(c *gin.Context) {
	var args entity.NodeNewTaskArgs
	if err := c.ShouldBindJSON(&args); err != nil {
		n.logger.Errorf("[node][NewTaskHandler] invalid args: %s", err.Error())
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
		return
	}
	// 测试URL是否是合法
	u, err := url.Parse(args.Target)
	if err != nil {
		n.logger.Error("[node][NewTaskHandler] invalid target url: %s", err.Error())
		c.JSON(200, gin.H{
			"success": false,
			"error":   "invalid target url",
			"data":    "",
		})
		return
	}
	//	[scheme:][//[userinfo@]host][/]path[?query][#fragment]
	uh := &url.URL{
		Scheme: u.Scheme,
		Opaque: u.Opaque,
		User:   u.User,
		Host:   u.Host,
	}
	if u.Path == "/" {
		u.Path = ""
	}
	cs := path.Dir(u.Path)
	if cs == "." {
		cs = ""
	}

	defaultSpider := entity.NewDefaultCrawlerConfig()
	// 根据用户配置的参数修改爬虫配置
	if args.SpiderConfig.CustomHeader != "" {
		defaultSpider.ExtraHeadersString = args.SpiderConfig.CustomHeader
	}
	if args.SpiderConfig.PostData != "" {
		defaultSpider.PostData = args.SpiderConfig.PostData
	}
	if args.SpiderConfig.MaxCrawledCount != 0 {
		defaultSpider.MaxCrawlCount = args.SpiderConfig.MaxCrawledCount
	}
	if args.SpiderConfig.MaxTabCount != 0 {
		defaultSpider.MaxTabsCount = args.SpiderConfig.MaxTabCount
	}
	if args.SpiderConfig.TabRunTimeout != 0 {
		defaultSpider.TabRunTimeout = time.Duration(args.SpiderConfig.TabRunTimeout)
	}
	if args.SpiderConfig.WaitDomContentLoadedTimeout != 0 {
		defaultSpider.DomContentLoadedTimeout = time.Duration(args.SpiderConfig.WaitDomContentLoadedTimeout)
	}
	if args.SpiderConfig.IgnoreURlKeywords != nil {
		defaultSpider.IgnoreKeywords = args.SpiderConfig.IgnoreURlKeywords
	}
	if args.SpiderConfig.FormValues != nil {
		defaultSpider.CustomFormValues = args.SpiderConfig.FormValues
	}
	if args.SpiderConfig.FormKeywordValues != nil {
		defaultSpider.CustomFormKeywordValues = args.SpiderConfig.FormKeywordValues
	}

	marshalSpiderConfig, err := json.Marshal(defaultSpider)
	if err != nil {
		n.logger.Errorf("[node][NewTaskHandler] marshal spider config failed: %s", err.Error())
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
		return
	}

	// 处理扫描参数
	//var scannerConfig entity.ScannerConfigs
	scannerConfig := args.ScannerConfig
	if args.ScannerConfig.RateLimit == 0 {
		scannerConfig.RateLimit = 300
	}
	if args.ScannerConfig.BulkSize == 0 {
		scannerConfig.BulkSize = 25
	}
	if args.ScannerConfig.Concurrency == 0 {
		scannerConfig.Concurrency = 50
	}

	marshalScannerConfig, err := json.Marshal(scannerConfig)
	if err != nil {
		n.logger.Errorf("[node][NewTaskHandler] marshal scanner config failed: %s", err.Error())
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
		return
	}

	newTask := entity.NodeTask{
		TaskID:        args.TaskID,
		TargetID:      args.TargetID,
		Target:        args.Target,
		TargetHost:    uh.String(),
		TargetPath:    u.Path,
		TargetPathDir: cs,
		SpiderConfig:  string(marshalSpiderConfig),
		ScannerConfig: string(marshalScannerConfig),
		Proxy:         "",
		ProxyType:     "",
		Status:        "created",
		ErrorReason:   "",
		Deleted:       false,
	}

	id, err := n.app.NewTask(&newTask)
	if err != nil {
		n.logger.Errorf("[node][NewTaskHandler] create new task failed: %s", err.Error())
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
	}

	n.logger.Infof("[node][NewTaskHandler] create new task success: %d", id)
	c.JSON(200, gin.H{
		"state": true,
		"msg":   "",
		"data":  "ok",
	})
}
