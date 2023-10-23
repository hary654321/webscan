package entity

import (
	"gorm.io/gorm"
	"strconv"
	"time"
)

// SlaveTask is a struct that represent the crawl task
type SlaveTask struct {
	ID          uint       `gorm:"primary_key;auto_increment" json:"id"`
	TaskID      uint       `gorm:"unsigned int;not null" json:"task_id"`
	TargetID    uint       `gorm:"unsigned int;not null" json:"target_id"`
	Target      string     `gorm:"text;not null" json:"target"`
	TargetHost  string     `gorm:"text;not null" json:"target_host"`
	TargetPath  string     `gorm:"text;not null" json:"target_path"`
	CrawlScope  string     `gorm:"text;not null" json:"crawl_scope"`
	Depth       uint       `gorm:"unsigned int;not null" json:"depth"`
	Timeout     uint       `gorm:"unsigned int;not null" json:"timeout"`
	Retry       uint       `gorm:"unsigned int;not null" json:"retry"`
	Proxy       string     `gorm:"text" json:"proxy"`
	ProxyType   string     `gorm:"char(16)" json:"proxy_type"`
	Status      string     `gorm:"char(16);not null" json:"status" comment:"new, crawling, scanning, finished, error"`
	ErrorReason string     `gorm:"text" json:"error_reason"`
	Deleted     bool       `gorm:"bool;not null" json:"deleted"`
	CreatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

func (t *SlaveTask) TableName() string {
	return "task"
}

func (t *SlaveTask) IDToString(id uint64) string {
	//uint64 to string
	return strconv.FormatUint(id, 10)
}

func (t *SlaveTask) BeforeSave(tx *gorm.DB) {
	if t.Depth == 0 {
		t.Depth = 3
	}
	if t.Timeout == 0 {
		t.Timeout = 10
	}
	if t.Retry == 0 {
		t.Retry = 3
	}
	if t.Status == "" {
		t.Status = "new"
	}
}
