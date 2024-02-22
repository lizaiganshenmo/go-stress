package node

import (
	"context"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/lizaiganshenmo/GoStress/cmd/stress/node/option"
	"github.com/lizaiganshenmo/GoStress/cmd/stress/task"
	"github.com/lizaiganshenmo/GoStress/library"
	"github.com/lizaiganshenmo/GoStress/resource"
)

type StressWorkDriver struct {
	NodeInfo
	// task chan
	taskCh chan *task.RequestTask
	// task map taskId-> RequestTaskWork
	taskMap map[int64]*task.RequestTaskWork

	taskMapMu *sync.RWMutex // 读写锁
}

// new StressWorkDriver
func NewStressWorkDriver(opts ...option.NodeInfoOption) *StressWorkDriver {
	return &StressWorkDriver{
		NodeInfo:  *NewNodeInfo(opts...),
		taskCh:    make(chan *task.RequestTask, 100),
		taskMap:   make(map[int64]*task.RequestTaskWork, 100),
		taskMapMu: &sync.RWMutex{},
	}

}

// work driver run
func (sw *StressWorkDriver) Run(ctx context.Context) {
	sw.Work(ctx)
}

// 发送节点信息
func (sw *StressWorkDriver) SendNodeInfo(ctx context.Context) {
	if sw.NodeInfo.NodeId == 0 {
		sw.NodeInfo = *NewNodeInfo()
	}

	// 节点信息发布
	sendStressNodeInfo(ctx, NodeInfoAdd, &sw.NodeInfo)

	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 节点信息更新
			sw.NodeInfo.UpdateNodeInfo(ctx)
			// 发布更新节点信息
			sendStressNodeInfo(ctx, NodeInfoModify, &sw.NodeInfo)
		}
	}

}

// 尝试成为master
func (sw *StressWorkDriver) TryBeMaster(ctx context.Context, sigCh chan string) {
	pubsub := resource.RedisCli.Subscribe(ctx, library.TryBeMasterChannel)
	defer pubsub.Close()

	for msg := range pubsub.Channel() {
		content := msg.Payload
		var info NodeInfo
		err := sonic.UnmarshalString(content, &info)
		if err != nil {
			hlog.Warnf("err happened. content: %s ; err: %+v ", content, err)
			continue
		}

		// 自己成为新的master了
		if info.NodeId == sw.NodeId {
			// sigCh <- struct{}{}
			sigCh <- StressMasterDriverName
			return
		}

	}

}

// 压力节点具体工作逻辑
func (sw *StressWorkDriver) Work(ctx context.Context) {
	// 接收任务
	go sw.receiveTask(ctx)
	// 处理任务
	go sw.handleTask(ctx)

}

// 接收任务
func (sw *StressWorkDriver) receiveTask(ctx context.Context) {
	pubsub := resource.RedisCli.Subscribe(ctx, library.GetNodeTaskChannelName(sw.NodeId))
	defer pubsub.Close()

	for msg := range pubsub.Channel() {
		content := msg.Payload
		var task task.RequestTask
		err := sonic.UnmarshalString(content, &task)
		if err != nil {
			hlog.Warnf("err happened. content: %s ; err: %+v ", content, err)
			continue
		}
		sw.taskCh <- &task
	}
}

// 处理任务
func (sw *StressWorkDriver) handleTask(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task1 := <-sw.taskCh:
			sw.taskMapMu.RLock()
			taskWork, exist := sw.taskMap[task1.TaskID]
			sw.taskMapMu.RUnlock()

			// 任务不存在
			if !exist {
				newCtx, cancelFunc := context.WithCancel(ctx)
				newTask := task.NewRequestTaskWork(newCtx, cancelFunc, task1)

				sw.taskMapMu.Lock()
				sw.taskMap[task1.TaskID] = newTask
				sw.taskMapMu.Unlock()

				// 启动任务
				newTask.Start()

			} else {
				// 任务停止
				if task1.TargetQPS == 0 {
					taskWork.Stop()

					sw.taskMapMu.Lock()
					delete(sw.taskMap, task1.TaskID)
					sw.taskMapMu.Unlock()
				} else {
					// 更改该任务QPS
					taskWork.ChangeQPS(task1.TargetQPS)
				}

			}

		}
	}
}
