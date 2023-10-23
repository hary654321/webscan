package v1

import (
	"github.com/gin-gonic/gin"
)

// SlaveTaskStatusHandler
// @Summary 获取Task的扫描进度
func (s *Slave) SlaveTaskStatusHandler(c *gin.Context) {
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

	status, err := s.app.GetTaskScanStatus(args.TaskID, args.TargetID)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
		return
	}
	s.logger.Infof("[SlaveTaskStatusHandler] task id: %d, target id: %d, status: %v", args.TaskID, args.TargetID, status)
	c.JSON(200, gin.H{
		"success": true,
		"error":   "",
		"data":    status,
	})
	return
}
