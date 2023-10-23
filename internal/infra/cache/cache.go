package cache

import (
	"sync"
)

type StateCacheInterface interface {
	// 由于一个任务对应多个目标，扫描节点以目标为单位
	// 所以，暂停的操作可以拆分为：将表中的任务状态(多条)置为paused，关闭爬虫与扫描器，清空缓存
	// 因此，对于缓存来说，只有 新建、查询、更新、删除 四个操作

	// New 创建一个新的状态缓存
	New(NodeTaskID uint)
	Get(NodeTaskID uint) *SpiderScanner
	UpdateSpiderFinishedState(NodeTaskID uint, req, allReq, allDomain, subDomain []string)
	UpdateScannerState(NodeTaskID uint, state string)
	Delete(NodeTaskID uint)
}

// StateCache 状态缓存
type StateCache struct {
	RWLock sync.RWMutex
	Status map[uint]*SpiderScanner
}

type SpiderScanner struct {
	SpiderStatus        string `json:"spider_status"` // running, finished
	SpiderStopChan      chan struct{}
	SpiderReqList       []string `json:"spider_req_list"`
	SpiderAllReqList    []string `json:"spider_all_req_list"`
	SpiderAllDomainList []string `json:"spider_all_domain_list"`
	SpiderSubDomainList []string `json:"spider_sub_domain_list"`

	ScannerStatus   string `json:"scanner_status"` // null, running, pause, finished
	ScannerStopChan chan struct{}
}

func NewStateCache() *StateCache {
	return &StateCache{
		RWLock: sync.RWMutex{},
		Status: make(map[uint]*SpiderScanner),
	}
}

var _ StateCacheInterface = &StateCache{}

// New 初始化一个新任务的状态
func (s *StateCache) New(NodeTaskID uint) {
	s.RWLock.Lock()
	defer s.RWLock.Unlock()
	s.Status[NodeTaskID] = &SpiderScanner{
		SpiderStatus:        "running",
		SpiderStopChan:      make(chan struct{}, 1),
		SpiderReqList:       make([]string, 0),
		SpiderAllReqList:    make([]string, 0),
		SpiderAllDomainList: make([]string, 0),
		SpiderSubDomainList: make([]string, 0),

		ScannerStatus:   "",
		ScannerStopChan: make(chan struct{}, 1),
	}
}

func (s *StateCache) Get(NodeTaskID uint) *SpiderScanner {
	s.RWLock.RLock()
	defer s.RWLock.RUnlock()
	if data, ok := s.Status[NodeTaskID]; ok {
		return data
	} else {
		return nil
	}
}

// UpdateSpiderFinishedState 爬虫执行完成后，设置缓存中爬虫执行的状态与返回结果
func (s *StateCache) UpdateSpiderFinishedState(NodeTaskID uint, req, allReq, allDomain, subDomain []string) {
	s.RWLock.Lock()
	defer s.RWLock.Unlock()
	if _, ok := s.Status[NodeTaskID]; ok {
		s.Status[NodeTaskID].SpiderStatus = "finished"
		s.Status[NodeTaskID].SpiderReqList = req
		s.Status[NodeTaskID].SpiderAllReqList = allReq
		s.Status[NodeTaskID].SpiderAllDomainList = allDomain
		s.Status[NodeTaskID].SpiderSubDomainList = subDomain
	}
}

func (s *StateCache) UpdateScannerState(NodeTaskID uint, state string) {
	s.RWLock.Lock()
	defer s.RWLock.Unlock()
	if _, ok := s.Status[NodeTaskID]; ok {
		s.Status[NodeTaskID].ScannerStatus = state
	}
}

func (s *StateCache) Delete(NodeTaskID uint) {
	s.RWLock.Lock()
	defer s.RWLock.Unlock()
	if _, ok := s.Status[NodeTaskID]; ok {
		close(s.Status[NodeTaskID].SpiderStopChan)
		close(s.Status[NodeTaskID].ScannerStopChan)
		delete(s.Status, NodeTaskID)
	}
}
