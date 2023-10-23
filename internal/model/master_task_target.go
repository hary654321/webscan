package model

import (
	"time"
)

// TaskTarget represents the 'task_target' table
type TaskTarget struct {
	ID         uint      `gorm:"column:id;type:int(10) unsigned;primaryKey;autoIncrement" json:"id"`
	TaskID     uint      `gorm:"column:task_id;type:int(11) unsigned" json:"task_id"`
	Target     string    `gorm:"column:target;type:varchar(255)" json:"target"`
	Describe   string    `gorm:"column:describe;type:varchar(255)" json:"describe"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;type:datetime" json:"updated_at"`
	IsColation uint      `gorm:"column:is_colation;type:int(11) unsigned;default:0" json:"is_colation"`
}

// TableName specifies the table name for TaskTarget struct
func (TaskTarget) TableName() string {
	return "task_target"
}
