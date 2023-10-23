package cache

import (
	"sync"
)

type NodeStatusInterface interface {
	Get(nodeIP string) *NodeStatus
	Set(nodeIP string, status NodeStatus)
	SetTaskStatus(nodeIP string, taskID, targetID uint)
	SetSystemStatus(nodeIP string, cpu, memTotal, memFree, memUsed, memUsedPercent string)
	Delete(nodeIP string)
	GetKeys() []string
}

type NodeStatusImpl struct {
	lock  *sync.Mutex
	cache map[string]NodeStatus
}

var _ NodeStatusInterface = &NodeStatusImpl{}

func NewNodeStatusCache() *NodeStatusImpl {
	return &NodeStatusImpl{
		lock:  &sync.Mutex{},
		cache: make(map[string]NodeStatus),
	}
}

func (n *NodeStatusImpl) Get(nodeIP string) *NodeStatus {
	n.lock.Lock()
	defer n.lock.Unlock()
	if v, ok := n.cache[nodeIP]; ok {
		return &v
	}
	return nil
}

func (n *NodeStatusImpl) Set(nodeIP string, ns NodeStatus) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.cache[nodeIP] = ns
}

func (n *NodeStatusImpl) SetTaskStatus(nodeIP string, taskID, targetID uint) {
	n.lock.Lock()
	defer n.lock.Unlock()

	if v, ok := n.cache[nodeIP]; ok {
		v.TaskID = taskID
		v.TargetID = targetID
		n.cache[nodeIP] = v
	} else {
		n.cache[nodeIP] = NodeStatus{
			Status:   "running",
			TaskID:   taskID,
			TargetID: targetID,
		}
	}
}

func (n *NodeStatusImpl) SetSystemStatus(nodeIP string, cpu, memTotal, memFree, memUsed, memUsedPercent string) {
	n.lock.Lock()
	defer n.lock.Unlock()

	if v, ok := n.cache[nodeIP]; ok {
		v.CPU = cpu
		v.MemTotal = memTotal
		v.MemFree = memFree
		v.MemUsed = memUsed
		v.MemUsedPercent = memUsedPercent
		n.cache[nodeIP] = v
	}
}

func (n *NodeStatusImpl) Delete(nodeIP string) {
	n.lock.Lock()
	defer n.lock.Unlock()
	delete(n.cache, nodeIP)
}

func (n *NodeStatusImpl) GetKeys() []string {
	n.lock.Lock()
	defer n.lock.Unlock()
	var keys []string
	for k := range n.cache {
		keys = append(keys, k)
	}
	return keys
}
