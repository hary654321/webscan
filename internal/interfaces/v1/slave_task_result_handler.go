package v1

import (
	"github.com/gin-gonic/gin"
)

// SlaveGetTaskResultHandler
// @Summary 获取target的扫描结果
func (s *Slave) SlaveGetTaskResultHandler(c *gin.Context) {
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
	s.logger.Infof("[SlaveGetTaskResultHandler] args: %+v", args)
	files, err := s.app.GetTaskScanResult(args.TaskID, args.TargetID)
	if err != nil {
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
