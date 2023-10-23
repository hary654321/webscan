package node

import (
	"github.com/gin-gonic/gin"
	"webscan/internal/domain/entity"
)

// DeleteTaskHandler
// @Summary 停止task执行
// @Description 停止task执行
// @Tags Node
// @Accept json
// @Produce json
// @Param args body entity.StopTaskRequestArgs
// @Success 200 {object} Response
// @Router /api/node/delete_task [post]
func (n *Node) DeleteTaskHandler(c *gin.Context) {
	var args entity.DeleteTaskRequestArgs
	err := c.BindJSON(&args)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
		return
	}

	tasks, err := n.app.FindTask(args.TaskID)
	if err != nil {
		n.logger.Errorf("[node][DeleteTaskHandler] find task error: %s", err.Error())
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
		return
	}

	n.app.DeleteTasks(tasks)

	n.logger.Infof("[node][DeleteTaskHandler] task id: %d", args.TaskID)

	c.JSON(200, gin.H{
		"success": true,
		"error":   "",
		"data":    "",
	})
	return
}
