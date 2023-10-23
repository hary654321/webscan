package v1

import (
	"github.com/gin-gonic/gin"
)

// SlaveNodeStatusHandler
// @Summary 获取当前节点的状态
func (s *Slave) SlaveNodeStatusHandler(c *gin.Context) {
	ns := s.app.GetNodeStatus()
	c.JSON(200, gin.H{
		"success": true,
		"error":   "",
		"data":    ns,
	})
	return
}
