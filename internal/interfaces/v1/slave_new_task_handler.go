package v1

import (
	"github.com/gin-gonic/gin"
	"net/url"
	"path"
	"webscan/internal/domain/entity"
)

// SlaveNewTaskHandler
// @Summary Create a new task
func (s *Slave) SlaveNewTaskHandler(c *gin.Context) {
	var args entity.NodeTask
	if err := c.ShouldBindJSON(&args); err != nil {
		s.logger.Errorf("[SlaveNewTaskHandler] invalid args: %s", err.Error())
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
		return
	}
	// test URL is a valid URL
	u, err := url.Parse(args.Target)
	if err != nil {
		s.logger.Error("[SlaveNewTaskHandler] valid url required")
		c.JSON(200, gin.H{
			"success": false,
			"error":   "valid url required",
			"data":    "",
		})
	}
	uh := &url.URL{
		Scheme: u.Scheme,
		Opaque: u.Opaque,
		User:   u.User,
		Host:   u.Host,
	}
	if u.Path == "/" {
		u.Path = ""
	}
	cs := path.Dir(u.Path)
	if cs == "." {
		cs = ""
	}

	newSlaveTask := entity.SlaveTask{
		TaskID:     args.TaskID,
		TargetID:   args.TargetID,
		Target:     args.Target,
		TargetHost: uh.String(),
		TargetPath: u.Path,
		CrawlScope: cs,
		Depth:      3,
		Timeout:    10,
		Status:     "new",
	}

	id, err := s.app.NewTask(&newSlaveTask)
	if err != nil {
		s.logger.Errorf("[SlaveNewTaskHandler] create new task failed: %s", err.Error())
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    "",
		})
	}

	s.logger.Infof("[SlaveNewTaskHandler] create new task success: %d", id)

	c.JSON(200, gin.H{
		"state": true,
		"msg":   "",
		"data":  "ok",
	})
}
