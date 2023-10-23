package entity

type PostTaskArgs struct {
	TaskID   uint   `json:"task_id" required:"true" comment:"任务ID"`
	TargetID uint   `json:"target_id" required:"true" comment:"目标ID"`
	Target   string `json:"target" required:"true" comment:"目标URL"`

	CrawlConfig struct {
		CustomHeaders               map[string]interface{} `json:"custom_headers" comment:"自定义HTTP头"`
		PostData                    string                 `json:"post_data" comment:"POST数据"`
		MaxCrawledCount             int                    `json:"max_crawled_count" comment:"爬虫最大任务数量，避免因伪静态造成长时间无意义抓取"`
		MaxTabCount                 int                    `json:"max_tab_count" comment:"爬虫同时开启最大标签页，即同时爬取的页面数量"`
		TabRunTimeout               int                    `json:"tab_run_timeout" comment:"单个Tab标签页的最大运行超时"`
		WaitDomContentLoadedTimeout int                    `json:"wait_dom_content_loaded_timeout" comment:"爬虫等待页面加载完毕的最大超时"`
		IgnoreURlKeywords           []string               `json:"ignore_url_keywords" comment:"不想访问的URL关键字，一般用于在携带Cookie访问时排除注销链接。logout exit"`
		FormValues                  map[string]string      `json:"form_values" comment:"自定义表单填充的值，按文本类型设置"`
		FormKeywordValues           map[string]string      `json:"form_keyword_values" comment:"自定义表单填充的值，按关键字模糊匹配设置"`
	} `json:"crawl_config" required:"true"`
}

type SpiderConfigs struct {
	CustomHeader                string            `json:"custom_headers" comment:"自定义HTTP头"`
	CustomCookie                string            `json:"custom_cookie" comment:"自定义cookie"`
	PostData                    string            `json:"post_data" comment:"POST数据"`
	MaxCrawledCount             int               `json:"max_crawled_count" comment:"爬虫最大任务数量，避免因伪静态造成长时间无意义抓取"`
	MaxTabCount                 int               `json:"max_tab_count" comment:"爬虫同时开启最大标签页，即同时爬取的页面数量"`
	TabRunTimeout               int               `json:"tab_run_timeout" comment:"单个Tab标签页的最大运行超时"`
	WaitDomContentLoadedTimeout int               `json:"wait_dom_content_loaded_timeout" comment:"爬虫等待页面加载完毕的最大超时"`
	IgnoreURlKeywords           []string          `json:"ignore_url_keywords" comment:"不想访问的URL关键字，一般用于在携带Cookie访问时排除注销链接。logout exit"`
	FormValues                  map[string]string `json:"form_values" comment:"自定义表单填充的值，按文本类型设置"`
	FormKeywordValues           map[string]string `json:"form_keyword_values" comment:"自定义表单填充的值，按关键字模糊匹配设置"`
}

type ScannerConfigs struct {
	CustomHeader    string `json:"custom_header" comment:"自定义HTTP头"`
	CustomCookie    string `json:"custom_cookie" comment:"自定义cookie"`
	RateLimit       int    `json:"rate_limit" comment:"每秒最大请求量（默认：150）"`
	RateLimitMinute int    `json:"rate_limit_minute" comment:"每分钟最大请求量"`
	BulkSize        int    `json:"bulk_size" comment:"每个模板最大并行检测数（默认：25）"`
	Concurrency     int    `json:"concurrency" comment:"并行执行的最大模板数量（默认：25）"`
}

type NodeNewTaskArgs struct {
	TaskID        uint           `json:"task_id" required:"true" comment:"任务ID"`
	TargetID      uint           `json:"target_id" required:"true" comment:"目标ID"`
	Target        string         `json:"target" required:"true" comment:"目标URL"`
	SpiderConfig  SpiderConfigs  `json:"spider_config" required:"true"`
	ScannerConfig ScannerConfigs `json:"scanner_config" required:"true"`
}

// StopTaskRequestArgs 停止任务请求参数
type StopTaskRequestArgs struct {
	TaskID   uint `json:"task_id" required:"true" comment:"任务ID"`
	TargetID uint `json:"target_id" required:"true" comment:"目标ID"`
}

type ResumeTaskRequestArgs struct {
	TaskID uint `json:"task_id" required:"true" comment:"任务ID"`
}

type DeleteTaskRequestArgs struct {
	TaskID uint `json:"task_id" required:"true" comment:"任务ID"`
}
