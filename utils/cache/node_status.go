package cache

type NodeStatus struct {
	Status   string `json:"status"` // "running" or "idle" or "offline"
	TaskID   uint   `json:"task_id"`
	TargetID uint   `json:"target_id"`
	CPU      string `json:"cpu"`

	MemTotal       string `json:"mem_total"`
	MemFree        string `json:"mem_free"`
	MemUsed        string `json:"mem_used"`
	MemUsedPercent string `json:"mem_used_percent"`
}
