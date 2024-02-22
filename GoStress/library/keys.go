package library

import "fmt"

const (
	HeartBeatRate       = 5  // 5s 心跳一下
	KeepAliveDuration   = 10 // 持续 10s
	MasterLockKey       = "stress_lock"
	NodeInfoChannel     = "node_info_channel"
	NodeTaskChannelPre  = "node_task_channel_"
	TryBeMasterChannel  = "be_master_channel" // 选举出谁是新的master节点，由该channel通知
	TaskDistributionKey = "task_distribution"
)

// 主节点向工作节点发送任务的channel名
func GetNodeTaskChannelName(nodeID int64) string {
	return fmt.Sprintf("%s%d", NodeTaskChannelPre, nodeID)
}
