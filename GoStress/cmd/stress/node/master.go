package node

import (
	"context"
)

type MasterDriver interface {
	BaseNodeDriver
	// 主节点保活
	KeepAlive(ctx context.Context)

	// 管理节点
	ManageNode(ctx context.Context)

	// 监听任务
	ListenTask(ctx context.Context)
	// 分发任务
	DistributeTask(ctx context.Context)
}
