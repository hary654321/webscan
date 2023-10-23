package node

import (
	"github.com/gin-gonic/gin"
)

// PullResultHandler
// @Summary 获取target的扫描结果
func (n *Node) PullResultHandler(c *gin.Context) {
	type b struct {
		TaskID   uint `json:"task_id"`
		TargetID uint `json:"target_id"`
	}
	var args b
	err := c.BindJSON(&args)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
		return
	}

	n.logger.Infof("[node][PullResultHandler] args: %+v", args)
	files, err := n.app.PullResult(args.TaskID, args.TargetID)
	if err != nil {
		n.logger.Errorf("[node][PullResultHandler] err: %s", err.Error())
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"error":   "",
		"data":    files,
	})
	return
}
