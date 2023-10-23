package model

import (
	"time"
)

// Scan represents the 'scan' table
type Scan struct {
	ID         uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TaskID     uint      `gorm:"column:task_id;type:int(11) unsigned" json:"task_id"`
	Name       string    `gorm:"column:name;type:varchar(255)" json:"name"`
	Node       string    `gorm:"column:node;type:varchar(255)" json:"node"`
	TargetID   uint      `gorm:"column:target_id;type:int(11) unsigned" json:"target_id"`
	ScanTarget string    `gorm:"column:scan_target;type:varchar(255)" json:"scan_target"`
	Status     uint      `gorm:"column:status;type:int(11);default:0" json:"status"`
	Pull       uint      `gorm:"column:pull;type:int(11);default:0" json:"pull"`
	StartTime  time.Time `gorm:"column:start_time;type:datetime" json:"start_time"`
	EndTime    time.Time `gorm:"column:end_time;type:datetime" json:"end_time"`
}

// TableName specifies the table name for Scan struct
func (Scan) TableName() string {
	return "scan"
}
