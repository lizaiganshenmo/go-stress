package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lizaiganshenmo/GoStress/cmd/stress/node"
	"github.com/lizaiganshenmo/GoStress/conf"
	"github.com/lizaiganshenmo/GoStress/library"
	"github.com/lizaiganshenmo/GoStress/resource"
)

var (
	isSentinel = 0
)

func init() {
	conf.Init("../../conf")
	// 加载全局变量
	resource.Init()

	// 压力测试节点驱动注册
	node.RegisterMasterDriver(library.StressMasterName, node.NewStressMasterDriver())
	node.RegisterWorkNodeDriver(library.StressNodeName, node.NewStressWorkDriver())

	// 程序入参
	flag.IntVar(&isSentinel, "sentinel", isSentinel, "默认不是哨兵节点")
}
func main() {
	var newNode *node.Node
	var err error
	if isSentinel > 0 {
		//成为哨兵
		fmt.Println("成为哨兵")

	} else {
		ok, err := resource.RedisCli.SetNX(context.Background(), library.MasterLockKey, 1, library.KeepAliveDuration).Result()
		if err != nil {
			panic(err)
		}

		if ok {
			// 成为master节点
			newNode, err = node.NewMasterNode(library.StressMasterName)
			if err != nil {
				panic(err)
			}
		} else {
			newNode, err = node.NewWorkNode(library.StressNodeName)
			if err != nil {
				panic(err)
			}
		}
	}
	if err != nil {
		panic(err)
	}
	newNode.Run()

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGINT, syscall.SIGTERM)

	<-stopSignal
	newNode.Stop()

}
