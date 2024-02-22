package node

import (
	"errors"
	"sort"
)

const (
	ExpectedNumber = 2 // 预期一个任务分配给两个节点
)

type StrategyFunc func(map[int64]*NodeInfo, int) (map[int64]int, error) // 根据nodes info ,targetQPS任务量 进行策略分发.  返回 m[nodeid1] = 100QPS ,m[nodeid2] = 200QPS

// 默认分配策略
// 根据节点的Cpu使用百分比/ 内存使用百分比/ 内存容量划分优先级
func defaultStrategy(nodeMap map[int64]*NodeInfo, targetQPS int) (map[int64]int, error) {

	if len(nodeMap) == 0 {
		return nil, errors.New("no have child node")
	}

	nList := make([]*NodeInfo, 0, len(nodeMap))
	for _, v := range nodeMap {
		nList = append(nList, v)
	}

	// 排序
	sort.Slice(nList, func(i, j int) bool {
		// 根据cpu占用升序排序
		if nList[i].CpuUsedPercent != nList[j].CpuUsedPercent {
			return nList[i].CpuUsedPercent < nList[j].CpuUsedPercent
		}

		// 根据内存占用升序排序
		if nList[i].MemoryUsedPercent != nList[j].MemoryUsedPercent {
			return nList[i].MemoryUsedPercent < nList[j].MemoryUsedPercent
		}

		// 根据内寸1容量降序排序
		if nList[i].Memory != nList[j].Memory {
			return nList[i].Memory > nList[j].Memory
		}

		return nList[i].NodeId < nList[j].NodeId
	})

	// todo 兜底策略，特定条件下(如只有一个节点且cpu使用率超100%，则不分配，以免影响其他正在执行任务)
	// 分配
	n := len(nList)
	if n == 1 {
		return map[int64]int{nList[0].NodeId: targetQPS}, nil
	}

	m := make(map[int64]int, ExpectedNumber)
	t := targetQPS / ExpectedNumber
	for i := 0; i < ExpectedNumber; i++ {
		m[nList[i].NodeId] = t
	}

	return m, nil
}
