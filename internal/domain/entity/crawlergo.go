package entity

import (
	"fmt"
	"github.com/Qianlitp/crawlergo/pkg"
	"github.com/Qianlitp/crawlergo/pkg/config"
	"time"
)

// Crawlergo 爬虫参数
type Crawlergo struct {
	AllDomainReturn         bool                   `json:"all_domain_return" comment:"全部域名收集"`
	SubDomainReturn         bool                   `json:"sub_domain_return" comment:"子域名收集"`
	CustomFormValues        map[string]string      `json:"custom_form_values" comment:"自定义表单填充参数"`
	ChromiumPath            string                 `json:"chromium_path" comment:"chrome的可执行程序路径"`
	ChromiumWSUrl           string                 `json:"chrome_ws_url" comment:"chromium websockets调试器的URL，注意：使用此选项时不会应用任何指定的chromium标志"`
	ExtraHeaders            map[string]interface{} `json:"extra_headers" comment:"chrome的可执行程序路径"`
	ExtraHeadersString      string                 `json:"extra_headers_string" comment:"自定义每个请求的HTTP头，输入字符串将进行json.Unmarshal"`
	PostData                string                 `json:"post_data" comment:"设置PostData到目标并使用POST请求方法"`
	MaxCrawlCount           int                    `json:"max_crawled_count" comment:"爬虫访问的最大URL数量，以避免因伪静态而导致的无效长时间抓取"`
	FilterMode              string                 `json:"filter_mode" comment:"用于收集请求的过滤模式。允许的模式：“simple”、“smart”或“strict”"`
	OutputMode              string                 `json:"output_mode" comment:"结果输出模式，console：打印当前域名结果。json：打印所有结果的JSON序列化字符串，可直接解析。none：不打印输出。"`
	OutputJsonPath          string                 `json:"output_json_path" comment:"将爬虫结果序列化为JSON后写入JSON文件"`
	MaxTabsCount            int                    `json:"max_tab_count" comment:"同时开启的最大标签页数量"`
	PathByFuzz              bool                   `json:"path_by_fuzz" comment:"是否使用常见路径Fuzz目标，以获取更多入口"`
	FuzzDictPath            string                 `json:"fuzz_dict_path" comment:"通过字典文件自定义Fuzz目录，传入字典文件路径，例如：/home/user/fuzz_dir.txt，文件的每一行代表一个要Fuzz的目录"`
	PathFromRobots          bool                   `json:"path_from_robots" comment:"是否从/robots.txt文件中解析路径，以获取更多入口"`
	Proxy                   string                 `json:"proxy" comment:"所有请求都通过定义的代理服务器连接"`
	EncodeURLWithCharset    bool                   `json:"encode_url_with_charset" comment:"是否使用检测到的字符集对URL进行编码"`
	TabRunTimeout           time.Duration          `json:"tab_run_timeout" comment:"单个标签页任务的最大运行超时"`
	DomContentLoadedTimeout time.Duration          `json:"dom_content_loaded_timeout" comment:"等待页面加载完毕的最大超时"`
	EventTriggerMode        string                 `json:"event_trigger_mode" comment:"事件自动触发的模式，分为异步和同步，用于处理DOM更新冲突导致的URL漏抓"`
	EventTriggerInterval    time.Duration          `json:"event_trigger_interval" comment:"触发每个事件的时间间隔"`
	BeforeExitDelay         time.Duration          `json:"before_exit_delay" comment:"在关闭Chrome之前等待单个标签页任务结束的时间，用于等待部分DOM更新和XHR请求的捕获"`
	IgnoreKeywords          []string               `json:"ignore_keywords" comment:"不希望访问的URL关键词列表，通常用于排除携带Cookie的注销链接"`
	CustomFormTypeValues    []string               `json:"custom_form_type_values" comment:"自定义表单填充的值，按文本类型设置"`
	CustomFormKeywordValues map[string]string      `json:"custom_form_keyword_values" comment:"自定义表单填充的值，按关键字模糊匹配设置"`
	PushAddress             string                 `json:"push_address" comment:"接收爬虫结果的监听地址"`
	PushProxyPoolMax        int                    `json:"push_proxy_pool_max" comment:"将爬虫结果发送到监听地址时的最大并发数"`
	LogLevel                string                 `json:"log_level" comment:"日志打印级别，可选项包括：debug、info、warn、error和fatal"`
	NoHeadless              bool                   `json:"no_headless" comment:"关闭Chrome的无头模式，可以直观地查看爬虫过程"`
	MaxRunTime              int64                  `json:"max_run_time" comment:"任务的最大运行超时"`
}

func NewDefaultCrawlerConfig() Crawlergo {
	return Crawlergo{
		AllDomainReturn:         false,
		SubDomainReturn:         false,
		CustomFormValues:        nil,
		ChromiumPath:            "",
		ChromiumWSUrl:           "",
		ExtraHeaders:            nil,
		ExtraHeadersString:      fmt.Sprintf(`{"User-Agent": "%s"}`, config.DefaultUA),
		PostData:                "",
		MaxCrawlCount:           config.MaxCrawlCount,
		FilterMode:              "smart",
		OutputMode:              "console",
		OutputJsonPath:          "",
		MaxTabsCount:            8,
		PathByFuzz:              false,
		FuzzDictPath:            "",
		PathFromRobots:          false,
		Proxy:                   "",
		EncodeURLWithCharset:    false,
		TabRunTimeout:           config.TabRunTimeout,
		DomContentLoadedTimeout: config.DomContentLoadedTimeout,
		EventTriggerMode:        config.EventTriggerAsync,
		EventTriggerInterval:    config.EventTriggerInterval,
		BeforeExitDelay:         config.BeforeExitDelay,
		IgnoreKeywords:          config.DefaultIgnoreKeywords,
		CustomFormTypeValues:    nil,
		CustomFormKeywordValues: nil,
		PushAddress:             "",
		PushProxyPoolMax:        10,
		LogLevel:                "info",
		NoHeadless:              false,
		MaxRunTime:              config.MaxRunTime,
	}
}

func (c *Crawlergo) NewTask() pkg.TaskConfig {
	t := pkg.TaskConfig{
		MaxCrawlCount:           c.MaxCrawlCount,
		FilterMode:              c.FilterMode,
		ExtraHeaders:            c.ExtraHeaders,
		ExtraHeadersString:      c.ExtraHeadersString,
		AllDomainReturn:         c.AllDomainReturn,
		SubDomainReturn:         c.SubDomainReturn,
		NoHeadless:              c.NoHeadless,
		DomContentLoadedTimeout: c.DomContentLoadedTimeout,
		TabRunTimeout:           c.TabRunTimeout,
		PathByFuzz:              c.PathByFuzz,
		FuzzDictPath:            c.FuzzDictPath,
		PathFromRobots:          c.PathFromRobots,
		MaxTabsCount:            c.MaxTabsCount,
		ChromiumPath:            c.ChromiumPath,
		ChromiumWSUrl:           c.ChromiumWSUrl,
		EventTriggerMode:        c.EventTriggerMode,
		EventTriggerInterval:    c.EventTriggerInterval,
		BeforeExitDelay:         c.BeforeExitDelay,
		EncodeURLWithCharset:    c.EncodeURLWithCharset,
		IgnoreKeywords:          c.IgnoreKeywords,
		Proxy:                   c.Proxy,
		CustomFormValues:        c.CustomFormValues,
		CustomFormKeywordValues: c.CustomFormKeywordValues,
		MaxRunTime:              c.MaxRunTime,
	}

	return t
}
