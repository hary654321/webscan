package entity

import "time"

/*
 task status
	1.created
	2.crawling
	3.crawled
	4.scanning
	5.finished
	6.error
	7.stopped
*/

// NodeTask 扫描任务
type NodeTask struct {
	ID                   uint      `gorm:"primary_key;auto_increment" json:"id"`
	TaskID               uint      `gorm:"unsigned int;not null" json:"task_id"`
	TargetID             uint      `gorm:"unsigned int;not null" json:"target_id"`
	Target               string    `gorm:"text;not null" json:"target"`
	TargetHost           string    `gorm:"text;not null" json:"target_host"`
	TargetPath           string    `gorm:"text;not null" json:"target_path"`
	TargetPathDir        string    `gorm:"text;not null" json:"target_path_dir"`
	SpiderConfig         string    `gorm:"text;not null" json:"spider_config"`
	ScannerConfig        string    `gorm:"text;not null" json:"scanner_config"`
	Proxy                string    `gorm:"char(256)" json:"proxy"`
	ProxyType            string    `gorm:"char(16)" json:"proxy_type"`
	Status               string    `gorm:"char(16);not null" json:"status" comment:"created, crawling, crawled, scanning, finished, error"`
	ErrorReason          string    `gorm:"text" json:"error_reason"`
	Deleted              bool      `gorm:"bool;not null" json:"deleted"`
	SpiderResultPullFlag bool      `json:"spider_result_pull_flag"`
	CreatedAt            time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt            time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (n *NodeTask) TableName() string {
	return "node_task"
}
