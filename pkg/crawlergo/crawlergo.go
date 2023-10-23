package crawlergo

import (
	"encoding/json"
	"github.com/Qianlitp/crawlergo/pkg"
	"github.com/Qianlitp/crawlergo/pkg/config"
	"github.com/Qianlitp/crawlergo/pkg/logger"
	"github.com/Qianlitp/crawlergo/pkg/model"
	"github.com/Qianlitp/crawlergo/pkg/tools"
	"github.com/google/uuid"
	"time"
	"webscan/internal/domain/entity"
)

// SpiderInterface 爬虫接口
type SpiderInterface interface {
	Start(target string, crawlergo *entity.Crawlergo) (SpiderResult, error)
	Stop(uuid4 uuid.UUID)
	State(uuid4 uuid.UUID)
}

type Spider struct{}

type SpiderResult struct {
	ReqList       []string `json:"req_list"`
	AllReqList    []string `json:"all_req_list"`
	AllDomainList []string `json:"all_domain_list"`
	SubDomainList []string `json:"sub_domain_list"`
}

func (s *Spider) Start(target string, spiderConfig *entity.Crawlergo) (SpiderResult, error) {
	var sr SpiderResult
	var targets []*model.Request

	var req model.Request
	url, err := model.GetUrl(target)
	if err != nil {
		return sr, err
	}

	var option model.Options
	if spiderConfig.PostData != "" {
		option.PostData = spiderConfig.PostData
	}
	if spiderConfig.ExtraHeadersString != "" {
		err := json.Unmarshal([]byte(spiderConfig.ExtraHeadersString), &spiderConfig.ExtraHeaders)
		if err != nil {
			return sr, err
		}
		option.Headers = spiderConfig.ExtraHeaders
	}

	if spiderConfig.PostData != "" {
		req = model.GetRequest(config.POST, url, option)
	} else {
		req = model.GetRequest(config.GET, url, option)
	}
	req.Proxy = spiderConfig.Proxy
	targets = append(targets, &req)

	//for _, _url := range c.Args().Slice() {
	//	var req model2.Request
	//	url, err := model2.GetUrl(_url)
	//	if err != nil {
	//		logger.Logger.Error("parse url failed, ", err)
	//		continue
	//	}
	//	if postData != "" {
	//		req = model.GetRequest(config.POST, url, getOption())
	//	} else {
	//		req = model.GetRequest(config.GET, url, getOption())
	//	}
	//	req.Proxy = taskConfig.Proxy
	//	targets = append(targets, &req)
	//}

	//taskConfig.IgnoreKeywords = ignoreKeywords.Value()
	//if taskConfig.Proxy != "" {
	//	logger.Logger.Info("request with proxy: ", taskConfig.Proxy)
	//}
	//
	//if len(targets) == 0 {
	//	logger.Logger.Fatal("no validate target.")
	//}

	//// 检查自定义的表单参数配置
	//taskConfig.CustomFormValues, err = parseCustomFormValues(customFormTypeValues.Value())
	//if err != nil {
	//	logger.Logger.Fatal(err)
	//}
	//taskConfig.CustomFormKeywordValues, err = keywordStringToMap(customFormKeywordValues.Value())
	//if err != nil {
	//	logger.Logger.Fatal(err)
	//}

	// 开始爬虫任务
	taskConfig := spiderConfig.NewTask()
	task, err := pkg.NewCrawlerTask(targets, taskConfig)
	if err != nil {
		logger.Logger.Error("create crawler task failed.")
		return sr, err
	}
	if len(targets) != 0 {
		logger.Logger.Infof("Init crawler task, "+
			"host: %s, "+
			"max tab count: %d, "+
			"max crawl count: %d, "+
			"max runtime: %ds, "+
			"filter mode: %s",
			targets[0].URL.Host, taskConfig.MaxTabsCount, taskConfig.MaxCrawlCount, taskConfig.MaxRunTime, taskConfig.FilterMode)
	}

	// 提示自定义表单填充参数
	if len(taskConfig.CustomFormValues) > 0 {
		logger.Logger.Info("Custom form values, " + tools.MapStringFormat(taskConfig.CustomFormValues))
	}
	// 提示自定义表单填充参数
	if len(taskConfig.CustomFormKeywordValues) > 0 {
		logger.Logger.Info("Custom form keyword values, " + tools.MapStringFormat(taskConfig.CustomFormKeywordValues))
	}
	if taskConfig.CustomFormValues != nil {
		if _, ok := taskConfig.CustomFormValues["default"]; !ok {
			logger.Logger.Info("If no matches, default form input text: " + config.DefaultInputText)
			taskConfig.CustomFormValues["default"] = config.DefaultInputText
		}
	}

	logger.Logger.Info("Start crawling.")
	task.Run()
	result := task.Result

	logger.Logger.Infof("Task finished, %d results, %d requests, %d subdomains, %d domains found, runtime: %d",
		len(result.ReqList), len(result.AllReqList), len(result.SubDomainList), len(result.AllDomainList), time.Now().Unix()-task.Start.Unix())

	var reqList []string
	var allReqList []string
	for _, _req := range result.ReqList {
		reqList = append(reqList, _req.URL.String())
	}
	for _, _req := range result.AllReqList {
		allReqList = append(allReqList, _req.URL.String())
	}
	sr.AllReqList = allReqList
	sr.ReqList = reqList
	sr.AllDomainList = result.AllDomainList
	sr.SubDomainList = result.SubDomainList

	return sr, nil
}

func (s *Spider) Stop(uuid4 uuid.UUID) {}

func (s *Spider) State(uuid4 uuid.UUID) {}

func NewSpider() *Spider {
	return &Spider{}
}
