package model

import (
	"time"
)

// Task represents the 'task' table
type Task struct {
	ID       uint   `gorm:"column:id;type:int(11) unsigned;primaryKey;autoIncrement" json:"id"`
	Name     string `gorm:"column:name;type:varchar(255)" json:"name"`
	Remark   string `gorm:"column:remark;type:varchar(255)" json:"remark"`
	TaskType uint   `gorm:"column:task_type;type:int(11) unsigned;default:0" json:"task_type"`
	Status   uint   `gorm:"column:status;type:int(11) unsigned;default:0" json:"status"`

	TargetType uint   `gorm:"column:target_type;type:int(11) unsigned;default:0" json:"target_type"`
	Target     string `gorm:"column:target;type:text" json:"target"`

	Bandwidth   uint `gorm:"column:bandwidth;type:int(11) unsigned;default:1024" json:"bandwidth"`
	Concurrency uint `gorm:"column:concurrency;type:int(11);default:0" json:"concurrency"`
	ScanType    uint `gorm:"column:scan_type;type:int(11) unsigned;default:0" json:"scan_type"`

	CreateTime time.Time `gorm:"column:create_time;type:datetime" json:"create_time"`
	StartTime  time.Time `gorm:"column:start_time;type:datetime" json:"start_time"`
	EndTime    time.Time `gorm:"column:end_time;type:datetime" json:"end_time"`
	UserToken  string    `gorm:"column:user_token;type:varchar(255)" json:"user_token"`

	IsCron           uint   `gorm:"column:is_cron;type:int(11);default:0" json:"is_cron"`
	IsRepeated       uint   `gorm:"column:is_repeated;type:int(11)" json:"is_repeated"`
	ScheduleStart    string `gorm:"column:schedule_start;type:varchar(255)" json:"schedule_start"`
	ScheduleEnd      string `gorm:"column:schedule_end;type:varchar(255)" json:"schedule_end"`
	DisableWhiteType string `gorm:"column:disable_white_type;type:varchar(255)" json:"disable_white_type"`
	Node             string `gorm:"column:node;type:varchar(255)" json:"node"`
	OnOff            uint   `gorm:"column:on_off;type:int(11) unsigned;default:1" json:"on_off"`
	Header           string `gorm:"column:header;type:text" json:"header"`
	Cookie           string `gorm:"column:cookie;type:text" json:"cookie"`
	IsExecute        uint   `gorm:"column:is_execute;type:int(11) unsigned;default:1" json:"is_execute"`

	TemplateID          uint   `gorm:"column:template_id;type:int(10) unsigned;default:0" json:"template_id"`
	RequestHeaders      string `gorm:"column:request_headers;type:text" json:"request_headers"`
	IsPreview           uint   `gorm:"column:is_preview;type:int(11) unsigned;default:0" json:"is_preview"`
	Describe            string `gorm:"column:describe;type:varchar(255)" json:"describe"`
	BusinessCriticality string `gorm:"column:business_criticality;type:varchar(255)" json:"business_criticality"`

	ScanSpeed           uint   `gorm:"column:scan_speed;type:int(10) unsigned;default:0" json:"scan_speed"`
	DetectionType       uint   `gorm:"column:detection_type;type:int(10) unsigned;default:0" json:"detection_type"`
	POCPluginFilePaths  string `gorm:"column:poc_plugin_file_paths;type:longtext" json:"poc_plugin_file_paths"`
	SecondMaxRequestNum uint   `gorm:"column:second_max_request_num;type:int(11) unsigned;default:0" json:"second_max_request_num"`
	ConcurrencyLevel    uint   `gorm:"column:concurrency_level;type:int(11) unsigned;default:0" json:"concurrency_level"`
	WithParamOnly       uint   `gorm:"column:with_param_only;type:int(11) unsigned;default:1" json:"with_param_only"`
	SecondTimeout       uint   `gorm:"column:second_timeout;type:int(11) unsigned;default:0" json:"second_timeout"`
	IsConfigEqually     uint   `gorm:"column:is_config_equally;type:int(11) unsigned;default:1" json:"is_config_equally"`
	// 爬虫配置
	SpiderConfig string `gorm:"column:spider_config;type:text" json:"spider_config"`
	// 扫描器配置
	ScannerConfig string `gorm:"column:scanner_config;type:text" json:"scanner_config"`

	TaskTargets []TaskTarget `gorm:"foreignKey:TaskID"`
}

// TableName specifies the table name for Task struct
func (Task) TableName() string {
	return "task"
}
