package node

//
//import (
//	"github.com/gin-gonic/gin"
//	"webscan/internal/domain/entity"
//)
//
//// StopTaskHandler
//// @Summary 停止task执行
//// @Description 停止task执行
//// @Tags Node
//// @Accept json
//// @Produce json
//// @Param args body entity.StopTaskRequestArgs
//// @Success 200 {object} Response
//// @Router /api/node/stop_task [post]
//func (n *Node) StopTaskHandler(c *gin.Context) {
//	var args entity.StopTaskRequestArgs
//	err := c.BindJSON(&args)
//	if err != nil {
//		c.JSON(200, gin.H{
//			"success": false,
//			"error":   err.Error(),
//			"data":    "",
//		})
//		return
//	}
//
//	task, err := n.app.FindOneTask(args.TaskID, args.TargetID)
//	if err != nil {
//		n.logger.Errorf("[node][StopTaskHandler] find task error: %s", err.Error())
//		c.JSON(200, gin.H{
//			"success": false,
//			"error":   err.Error(),
//			"data":    "",
//		})
//		return
//	}
//
//	err = n.app.StopTask(&task)
//	if err != nil {
//		n.logger.Errorf("[node][StopTaskHandler] stop task error: %s", err.Error())
//		c.JSON(200, gin.H{
//			"success": false,
//			"error":   err.Error(),
//			"data":    "",
//		})
//		return
//	}
//
//	n.logger.Infof("[node][StopTaskHandler] id: %s, task id: %d, target id: %d, target: %s",
//		task.ID, task.TaskID, task.TargetID, task.Target)
//
//	c.JSON(200, gin.H{
//		"success": true,
//		"error":   "",
//		"data":    "",
//	})
//	return
//}
