package node

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/lizaiganshenmo/GoStress/cmd/stress/node/option"
	"github.com/lizaiganshenmo/GoStress/cmd/stress/task"
	"github.com/lizaiganshenmo/GoStress/library"
	"github.com/lizaiganshenmo/GoStress/resource"
	"github.com/lizaiganshenmo/GoStress/resource/db"
)

type StressMasterDriver struct {
	NodeInfo
	// 子节点列表
	childNodes map[int64]*NodeInfo // key : nodeId
	// task channel
	taskCh chan *task.RequestTask

	// 任务量分发策略
	distributeStrategy StrategyFunc
	// work driver 支持单机执行压测任务
	WorkDriver *StressWorkDriver
}

// new StressMasterDriver
func NewStressMasterDriver(opts ...option.NodeInfoOption) *StressMasterDriver {
	nodeInfo := *NewNodeInfo(opts...)
	return &StressMasterDriver{
		NodeInfo:           nodeInfo,
		childNodes:         make(map[int64]*NodeInfo, 100),
		taskCh:             make(chan *task.RequestTask, 100),
		distributeStrategy: defaultStrategy,
		WorkDriver:         NewStressWorkDriver(option.WithNodeID(nodeInfo.NodeId)),
	}

}

// driver run
func (sm *StressMasterDriver) Run(ctx context.Context) {
	// 支持单机执行压测任务
	go sm.WorkDriver.Run(ctx)
	// 主节点功能执行
	go sm.KeepAlive(ctx)
	go sm.ManageNode(ctx)
	go sm.ListenTask(ctx)
	go sm.DistributeTask(ctx)
}

// 主节点信息
func (sm *StressMasterDriver) SendNodeInfo(ctx context.Context) {
	if sm.NodeInfo.NodeId == 0 {
		sm.NodeInfo = *NewNodeInfo()
	}

	// 节点信息发布
	sendStressNodeInfo(ctx, NodeInfoAdd, &sm.NodeInfo)

	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 节点信息更新
			sm.NodeInfo.UpdateNodeInfo(ctx)
			// 发布更新节点信息
			sendStressNodeInfo(ctx, NodeInfoModify, &sm.NodeInfo)
		}
	}

}

// 压力测试节点发送节点信息
func sendStressNodeInfo(ctx context.Context, cmdType int, info *NodeInfo) {
	cmdStr, _ := sonic.MarshalString(
		InfoCommand{
			Command: cmdType,
			Info:    info,
		})

	err := resource.RedisCli.Publish(ctx, library.NodeInfoChannel, cmdStr).Err()
	if err != nil {
		hlog.Warnf("RedisCli.Publish failed. err: %+v", err)
	}

}

// 尝试成为主节点
func (sm *StressMasterDriver) TryBeMaster(ctx context.Context, sigCh chan string) {
}

// 主节点保活
func (sm *StressMasterDriver) KeepAlive(ctx context.Context) {
	ticker := time.NewTicker(library.HeartBeatRate * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := resource.RedisCli.Expire(ctx, library.MasterLockKey, library.KeepAliveDuration*time.Second).Err()
			if err != nil {
				hlog.Warnf("RedisCli.Expire fail. err:%+v", err)
			}
		}
	}

}

// 主节点管理工作节点
func (sm *StressMasterDriver) ManageNode(ctx context.Context) {
	pubsub := resource.RedisCli.Subscribe(ctx, library.NodeInfoChannel)
	defer pubsub.Close()

	for msg := range pubsub.Channel() {
		content := msg.Payload
		var cmd InfoCommand
		err := sonic.UnmarshalString(content, &cmd)
		if err != nil {
			hlog.Warnf("err happened. content: %s ; err: %+v ", content, err)
			continue
		}

		sm.handleNodeInfo(&cmd)

	}
}

// 处理节点信息
func (sm *StressMasterDriver) handleNodeInfo(cmd *InfoCommand) {

	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	switch cmd.Command {
	case NodeInfoAdd, NodeInfoModify:
		sm.childNodes[cmd.Info.NodeId] = cmd.Info
	case NodeInfoDelete:
		delete(sm.childNodes, cmd.Info.NodeId)
	}
}

// 监听任务,任务传至Channel tashCh
func (sm *StressMasterDriver) ListenTask(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ctx.Done():
			hlog.Infof("receive ctx done signal. sm.ListenTask exit.")
			return
		case <-ticker.C:
			tasks, err := db.GetTaskByStatuss(ctx, []int{db.WaitExec, db.WaitChangeQPS, db.WaitStop})
			if err != nil {
				hlog.Warnf("db.GetNoExecTask fail. err:%+v", err)
			}

			for _, v := range tasks {
				tmp := v
				reqTask := task.FormatRequestTask(tmp)
				if reqTask == nil {
					hlog.Warnf("invalid tmp task :%+v", tmp.TaskID)
				}

				sm.taskCh <- reqTask
			}
		}
	}

}

// 分发任务
func (sm *StressMasterDriver) DistributeTask(ctx context.Context) {
	for v := range sm.taskCh {
		sm.handleTask2WorkNode(ctx, v)
	}

}

// 处理一个任务 经过策略分发，任务消息传至对应的node节点
func (sm *StressMasterDriver) handleTask2WorkNode(ctx context.Context, reqTask *task.RequestTask) {
	// 跟新db该任务状态
	err := db.UpdateTask(ctx, &db.StressTask{
		TaskID: reqTask.TaskID,
		Status: db.Execing,
	})
	if err != nil {
		hlog.Warnf("db.UpdateTask fail. err : %+v", err)
		return
	}

	hash, err := sm.handleTaskStatus(ctx, reqTask)
	if err != nil {
		hlog.Warnf("sm.handleTaskStatus fail. err : %+v", err)
		return
	}

	for nodeID, QPS := range hash {
		// 发送对任务
		id := nodeID
		qps := QPS
		task1 := *reqTask
		task1.TargetQPS = qps

		go func(int64, *task.RequestTask) {
			// todo 任务分发异常操作，如是否要增加工作节点接收任务机制, 发送失败重试或工作节点健康探测、重新分配节点
			err := sm.sendTask(ctx, id, &task1)
			if err != nil {
				hlog.Warnf("sm.sendTask fail. nodeid,task:%d,%+v. err: %+v", id, task1, err)
			}
		}(id, &task1)

	}

}

// 根据任务状态处理
func (sm *StressMasterDriver) handleTaskStatus(ctx context.Context, reqTask *task.RequestTask) (hash map[int64]int, err error) {
	taskIDStr := strconv.FormatInt(reqTask.TaskID, 10)
	switch reqTask.TaskStatus {
	case db.WaitExec:
		sm.Mu.RLock()
		hash, err = sm.distributeStrategy(sm.childNodes, reqTask.TargetQPS)
		sm.Mu.RUnlock()
		if err != nil {
			hlog.Warnf("sm.distributeStrategy fail. err : %+v.  sm.childNodes, reqTask.TargetQPS:%+v,%d", err, sm.childNodes, reqTask.TargetQPS)
			return
		}

		// 存储该任务分配详情
		err = sm.setTaskDistribution(ctx, taskIDStr, hash)
		if err != nil {
			hlog.Warnf("sm.setTaskDistribution fail. err : %+v.", err)
		}
	case db.WaitChangeQPS:
		hash, err = sm.getTaskDistribution(ctx, taskIDStr)
		if err != nil {
			hlog.Warnf("sm.getTaskDistribution fail. err : %+v.", err)
		}

		// todo 此处为更改各任务节点新QPS逻辑，后续应当添加策略优化
		preTotal := 0
		for _, v := range hash {
			preTotal += v
		}

		preTotalFloat := float64(preTotal)
		targetFloat := float64(reqTask.TargetQPS)
		for k, v := range hash {
			hash[k] = int(targetFloat * (float64(v) / preTotalFloat))
		}

		// 存储该任务分配详情
		err = sm.setTaskDistribution(ctx, taskIDStr, hash)
		if err != nil {
			hlog.Warnf("sm.setTaskDistribution fail. err : %+v.", err)
		}

	case db.WaitStop:
		hash, err = sm.getTaskDistribution(ctx, taskIDStr)
		if err != nil {
			hlog.Warnf("sm.getTaskDistribution fail. err : %+v.", err)
		}
		// 删除该任务分配详情
		err = sm.delTaskDistribution(ctx, taskIDStr)
		if err != nil {
			hlog.Warnf("sm.getTaskDistribution fail. err : %+v.", err)
		}

		for k := range hash {
			hash[k] = 0 //各节点的targetQPS更改为0
		}

		err = db.UpdateTask(ctx, &db.StressTask{
			TaskID: reqTask.TaskID,
			Status: db.HasExeced,
		})
		if err != nil {
			hlog.Warnf("db.UpdateTask fail. err : %+v", err)
			return
		}

	default:
		err = errors.New("invalid task status")
	}

	return

}

// 发送任务到对应的通道节点
func (sm *StressMasterDriver) sendTask(ctx context.Context, nodeID int64, reqTask *task.RequestTask) error {
	val, _ := sonic.MarshalString(reqTask)

	err := resource.RedisCli.Publish(ctx, library.GetNodeTaskChannelName(nodeID), val).Err()
	if err != nil {
		hlog.Warnf("sendTask RedisCli.Publish failed. err: %+v", err)
	}

	return err

}

// 获取某个已分配任务的分配情况
func (sm *StressMasterDriver) getTaskDistribution(ctx context.Context, taskIDStr string) (map[int64]int, error) {
	res, err := resource.RedisCli.HGet(ctx, library.TaskDistributionKey, taskIDStr).Result()
	if err != nil {
		return nil, err
	}

	var hash map[int64]int
	err = sonic.UnmarshalString(res, &hash)
	return hash, err
}

// 设置某个已分配任务的分配情况
func (sm *StressMasterDriver) setTaskDistribution(ctx context.Context, taskIDStr string, hash map[int64]int) error {
	hashStr, _ := sonic.MarshalString(hash)
	return resource.RedisCli.HSet(ctx, library.TaskDistributionKey, taskIDStr, hashStr).Err()
}

// 删除某个已分配任务的分配情况
func (sm *StressMasterDriver) delTaskDistribution(ctx context.Context, taskIDStr string) error {
	return resource.RedisCli.HDel(ctx, library.TaskDistributionKey, taskIDStr).Err()
}
