package model

import (
	"time"
)

// Node represents the 'node' table
type Node struct {
	ID         uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Address    string    `gorm:"column:address;type:varchar(255)" json:"address"`
	Remark     string    `gorm:"column:remark;type:varchar(255)" json:"remark"`
	VPSType    uint      `gorm:"column:vps_type;type:int(11);default:0" json:"vps_type"`
	TaskCount  int64     `gorm:"column:task_count;type:bigint(20);default:0" json:"task_count"`
	CPU        string    `gorm:"column:cpu;type:varchar(255)" json:"cpu"`
	Disk       string    `gorm:"column:disk;type:varchar(255)" json:"disk"`
	Mem        string    `gorm:"column:mem;type:varchar(255)" json:"mem"`
	Effect     uint      `gorm:"column:effect;type:int(11);default:1" json:"effect"`
	Online     uint      `gorm:"column:online;type:int(11)" json:"online"`
	OnlineTime time.Time `gorm:"column:online_time;type:datetime" json:"online_time"`
	Creator    string    `gorm:"column:creator;type:varchar(255);comment:创建者" json:"creator"`
	Unit       string    `gorm:"column:unit;type:varchar(255);comment:创建者单位信息" json:"unit"`
}

// TableName specifies the table name for Node struct
func (Node) TableName() string {
	return "node"
}
