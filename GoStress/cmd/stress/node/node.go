package node

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/lizaiganshenmo/GoStress/cmd/stress/node/option"
	"github.com/lizaiganshenmo/GoStress/library/utils"
	"github.com/lizaiganshenmo/GoStress/resource"
	"golang.org/x/sync/errgroup"
)

const (
	NodeMaster = iota
	NodeWork
	NodeSentinel
)

var (
	masterMu      sync.RWMutex
	masterDrivers = make(map[string]MasterDriver)

	workMu      sync.RWMutex
	workDrivers = make(map[string]WorkNodeDriver)
)

// 注册主节点driver
func RegisterMasterDriver(name string, master MasterDriver) {
	masterMu.Lock()
	defer masterMu.Unlock()

	if master == nil {
		panic("RegisterMaster master is nil")
	}

	if _, dup := masterDrivers[name]; dup {
		panic("RegisterMaster master called twice for driver " + name)
	}

	masterDrivers[name] = master

}

// 注册工作节点driver
func RegisterWorkNodeDriver(name string, driver WorkNodeDriver) {
	workMu.Lock()
	defer workMu.Unlock()

	if driver == nil {
		panic("RegisterWorkNodeDriver node is nil")
	}

	if _, dup := workDrivers[name]; dup {
		panic("RegisterWorkNodeDriver  called twice for driver " + name)
	}

	workDrivers[name] = driver

}

type BaseNodeDriver interface {
	Run(ctx context.Context)                            // driver运行
	SendNodeInfo(ctx context.Context)                   // 发送节点信息
	TryBeMaster(ctx context.Context, sigCh chan string) // 尝试成为master节点
}

type Node struct {
	Ctx      context.Context //上下文
	NodeType int             // 节点类型 0:主节点 1:工作节点 2:哨兵
	// Info   *NodeInfo
	Driver BaseNodeDriver // 类似实际驱动, 执行master、 work node 、 sentinel node 实际功能
	stop   func()         // 节点停止
}

// 尝试变为主节点
func (n *Node) tryBeMasterNode(ctx context.Context) {
	if n.Driver == nil {
		return
	}

	sigCh := make(chan string)
	go n.Driver.TryBeMaster(ctx, sigCh)

	for {
		select {
		case <-ctx.Done():
			return
		case masterName := <-sigCh:
			hlog.Infof("receive be master node signal. now change to master.")
			// 成为master
			n.beMaster(ctx, masterName)
		}
	}
}

// 成为master
func (n *Node) beMaster(ctx context.Context, masterName string) {
	n.NodeType = NodeMaster
	// 节点驱动停止
	n.Stop()

	masterMu.RLock()
	masterDriver, ok := masterDrivers[masterName]
	masterMu.RUnlock()
	if !ok {
		hlog.Warnf("invalid master driver name: %s .", masterName)
		panic(fmt.Sprintf("invalid master driver name: %s .", masterName))
	}

	n.Driver = masterDriver
	go n.Driver.SendNodeInfo(ctx)

}

// node run
func (n *Node) Run() {
	n.Driver.Run(n.Ctx)
}

// node stop
func (n *Node) Stop() {
	n.stop()
}

// 创建主节点
func NewMasterNode(name string) (*Node, error) {
	masterMu.RLock()
	masterDriver, ok := masterDrivers[name]
	masterMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("invalid master driver name: %s ", name)
	}

	return newNode(masterDriver, NodeMaster), nil

}

// 创建工作节点
func NewWorkNode(name string) (*Node, error) {
	workMu.RLock()
	workDriver, ok := workDrivers[name]
	workMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("invalid work driver name: %s ", name)
	}

	return newNode(workDriver, NodeWork), nil

}

// 创建节点
func newNode(d BaseNodeDriver, nodeType int) *Node {
	ctx, cancel := context.WithCancel(context.Background())
	node := Node{
		Ctx:      ctx,
		NodeType: nodeType,
		// Info:   NewNodeInfo(),
		Driver: d,
		stop:   cancel,
	}

	// 持续定时更新节点信息
	go node.Driver.SendNodeInfo(ctx)

	// 如果是工作节点,持续监听自己是否被选举为master节点
	if nodeType == NodeWork {
		go node.tryBeMasterNode(ctx)
	}

	return &node

}

// 节点信息
type NodeInfo struct {
	Host              string        // 节点主机ip
	NodeId            int64         // 节点id
	Memory            uint64        // 内存容量
	CpuUsedPercent    float64       // CPU占用百分比
	MemoryUsedPercent float64       // 内存占用百分比
	Mu                *sync.RWMutex // 读写锁
	CreateAt          time.Time     // 创建时间
	UpdateAt          time.Time     // 更新时间
}

// 更新节点信息
func (n *NodeInfo) UpdateNodeInfo(ctx context.Context) {
	eg, _ := errgroup.WithContext(ctx)
	var memTotal uint64
	var memUsedPercent float64
	var ip string
	var cpuPercent float64
	now := time.Now()

	eg.Go(
		func() error {
			memTotal, memUsedPercent = utils.GetMemoryInfo()
			return nil
		},
	)
	eg.Go(
		func() error {
			ip = utils.GetOutboundIP()
			return nil
		},
	)
	eg.Go(
		func() error {
			cpuPercent = utils.GetCPUPercent()
			return nil
		},
	)
	if err := eg.Wait(); err != nil {
		hlog.Warnf("UpdateNodeInfo  err happens. err:%+v\n", err)
	}

	n.Mu.Lock()
	n.Host = ip
	n.Memory = memTotal
	n.MemoryUsedPercent = memUsedPercent
	n.CpuUsedPercent = cpuPercent
	n.UpdateAt = now
	n.Mu.Unlock()
}

// 创建节点信息
func NewNodeInfo(opts ...option.NodeInfoOption) *NodeInfo {
	o := option.NewNodeInfoOptions(opts)
	nodeID := o.NodeId
	if nodeID == 0 {
		nodeID = resource.SF.NextVal()
	}
	now := time.Now()
	n := NodeInfo{
		Host:     o.Host,
		NodeId:   nodeID,
		Mu:       &sync.RWMutex{},
		CreateAt: now,
		UpdateAt: now,
	}

	return &n
}

const (
	NodeInfoAdd = iota
	NodeInfoModify
	NodeInfoDelete
)

// 发送节点信息 命令结构
type InfoCommand struct {
	Command int // NodeInfoAdd...
	Info    *NodeInfo
}
