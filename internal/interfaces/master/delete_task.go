package master

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"webscan/internal/domain/entity"
)

// DeleteTaskHandler
// @Router /api/master/delete_task [post]
func (m *Master) DeleteTaskHandler(c *gin.Context) {
	var args RequestArgs
	err := c.BindJSON(&args)
	if err != nil {
		m.logger.Errorf("[master][DeleteTaskHandler] bind json error: %s", err.Error())
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
		return
	}

	var nodeResumeTask entity.ResumeTaskRequestArgs
	nodeResumeTask.TaskID = args.TaskID
	resp, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(nodeResumeTask).
		Post(fmt.Sprintf("http://%s:18080/api/node/v1/delete_task", args.NodeIP))

	if err != nil {
		m.logger.Errorf("[master][DeleteTaskHandler] request node error: %s", err.Error())
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
		return
	}

	var respBody map[string]interface{}
	err = json.Unmarshal(resp.Body(), &respBody)
	if err != nil {
		m.logger.Errorf("[master][DeleteTaskHandler] unmarshal response body error: %s", err.Error())
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
		return
	}

	if !respBody["success"].(bool) {
		m.logger.Errorf("[master][DeleteTaskHandler] request node error: %s", respBody["error"].(string))
		c.JSON(200, gin.H{
			"success": false,
			"error":   respBody["error"].(string),
			"data":    "",
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"error":   "",
		"data":    "",
	})
	return
}
