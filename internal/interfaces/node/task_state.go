package node

import (
	"github.com/gin-gonic/gin"
	"webscan/internal/domain/entity"
)

// TaskStateHandler
// @Summary 获取任务状态
func (n *Node) TaskStateHandler(c *gin.Context) {
	var args entity.StopTaskRequestArgs
	if err := c.ShouldBindJSON(&args); err != nil {
		n.logger.Errorf("[TaskState] invalid args: %s", err.Error())
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
		return
	}

	task, err := n.app.FindOneTask(args.TaskID, args.TargetID)
	if err != nil {
		n.logger.Errorf("[TaskState] invalid task: %s", err.Error())
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
		"data":    task.Status,
	})
	return
}
