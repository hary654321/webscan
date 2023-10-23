package node

import (
	"github.com/gin-gonic/gin"
	"webscan/internal/domain/entity"
)

// PauseTaskHandler
// @Summary 暂停task执行
// @Description 暂停task执行
// @Tags Node
// @Accept json
// @Produce json
// @Param args body entity.StopTaskRequestArgs
// @Success 200 {object} Response
// @Router /api/node/pause_task [post]
func (n *Node) PauseTaskHandler(c *gin.Context) {
	var args entity.StopTaskRequestArgs
	err := c.BindJSON(&args)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
		return
	}

	err = n.app.PauseTask(args.TaskID)
	if err != nil {
		n.logger.Errorf("[node][PauseTaskHandler] pause task error: %s", err.Error())
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
		return
	}

	n.logger.Infof("[node][PauseTaskHandler] task id: %d", args.TaskID)

	c.JSON(200, gin.H{
		"success": true,
		"error":   "",
		"data":    "",
	})
	return
}
