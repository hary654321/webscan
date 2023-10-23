package node

import (
	"github.com/gin-gonic/gin"
	"webscan/internal/domain/entity"
)

// ResumeTaskHandler
// @Summary 恢复暂停的task继续执行
// @Description 恢复暂停的task继续执行
// @Tags Node
// @Accept json
// @Produce json
// @Param args body entity.StopTaskRequestArgs
// @Success 200 {object} Response
// @Router /api/node/resume_task [post]
func (n *Node) ResumeTaskHandler(c *gin.Context) {
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

	err = n.app.ResumeTask(args.TaskID)
	if err != nil {
		n.logger.Errorf("[node][ResumeTaskHandler] resume task error: %s", err.Error())
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
		return
	}

	n.logger.Infof("[node][ResumeTaskHandler] task id: %d", args.TaskID)

	c.JSON(200, gin.H{
		"success": true,
		"error":   "",
		"data":    "",
	})
	return
}
