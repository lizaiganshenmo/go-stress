## GoStress初步设计草图

- ### 整体草图

  <img src="https://github.com/lizaiganshenmo/pictures/blob/main/%E6%80%BB%E4%BD%93%E6%A6%82%E8%A7%88.png" alt="总体概览" style="zoom: 50%;" />

  任务中心

  <img src="https://github.com/lizaiganshenmo/pictures/blob/main/%E5%88%86%E5%B8%83%E5%BC%8F%E4%BB%BB%E5%8A%A1%E8%8A%82%E7%82%B9.png" alt="分布式任务节点" style="zoom:33%;" />

- ### 节点介绍

  - #### Master主节点选举

    1. 初步通过抢占redis分布式锁当选主节点；
    2. master节点宕机（哨兵投票判断客观下线）后，由哨兵（获得超半数赞成票的哨兵）执行选主；
    3. 选主细节：过滤+打分选择。 根据cpu/内存等情况，初步过滤；根据节点优先级、节点cpu情况、节点序号打分选择。

  - #### 主节点基础特性

    ```go
    type BaseNodeDriver interface {
    	Run(ctx context.Context)                            // driver运行
    	SendNodeInfo(ctx context.Context)                   // 发送节点信息
    	TryBeMaster(ctx context.Context, sigCh chan string) // 尝试成为master节点
    }
    ```

    

    ```go
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
    ```

    

  - #### 工作节点基础特性

    ```go
    type WorkNodeDriver interface {
    	BaseNodeDriver
    	Work(ctx context.Context)
    }
    
    ```

    

  - #### 哨兵节点

    **TODO**，哨兵节点暂未开发，哨兵节点应当具有如下特性：

    1. 监控主节点及工作节点健康状态；
    2. 完善的投票选主机制。

  

- ### 设计详情

  1. #### 节点功能

     ​        除了基本的节点基础功能(如发送自身节点信息...），节点的实际额外功能属性，通过节点自身的driver具体执行。如master节点额外逻辑实际是在MasterDriver的具体对象上。

     ```go
     type Node struct {
     	Ctx      context.Context //上下文
     	NodeType int             // 节点类型 0:主节点 1:工作节点 2:哨兵
     	// Info   *NodeInfo
     	Driver BaseNodeDriver // 类似实际驱动, 执行master、 work node 、 sentinel node 实际功能
     	stop   func()         // 节点停止
     }
     ```

     ```go
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
     ```

     ​        相当于提供一套分布式任务通用模板，而用户需要自定义实现自己的MasterDriver，WorkNodeDriver，并通过node对外提供的RegisterDriver来注册驱动。

     ```go
     // 压力测试节点驱动注册
     	node.RegisterMasterDriver(library.StressMasterName, node.NewStressMasterDriver())
     	node.RegisterWorkNodeDriver(library.StressNodeName, node.NewStressWorkDriver())
     ```

     ​       此处设计，参考golang版sql标准库源码。http://go-database-sql.org/index.html  、 https://github.com/go-sql-driver/mysql

  2. #### 压力测试主节点驱动Driver

     ```go
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
     ```

     - ##### KeepAlive心跳

       通过定时向Redis的nx key续期，来实现主节点心跳保活功能；

     - ##### 发送节点自身信息

       定时向特定redis channel发送自身节点信息；

     - ##### 管理工作节点

       订阅redis 节点信心channel，对节点信息的新增、修改、删除等，存在sm.childNodes下；

     - ##### 监听任务

       定时从db读取待执行、待更改QPS、待停止状态的任务,通过taskCh进行消费；

     - ##### 分发任务

       从taskCh读取任务，根据任务不同状态进行特定处理：

       ​        待执行状态：通过任务分配策略，从子节点列表中分配出一定节点，各个节点执行对应分配的QPS target；

       ​        待更改QPS状态：读取该任务之前分配的节点，按照之前分配的QPS比重，重新将新的目标QPS分配给对应节点；

       ​        待停止状态：读取该任务之前分配的节点，任务QPS清零，对应任务信息删除；

       任务发送，通过特定redis管道发送该条任务，工作节点应当订阅自己的任务管道；

     - ##### 主节点自身执行压测任务

       为了支持单机压测任务，主节点同时也要支持执行压测任务，所以StressMasterDriver下新增WorkDriver；

  3. #### 压力测试工作节点驱动Driver

     ```go
     type StressWorkDriver struct {
     	NodeInfo
     	// task chan
     	taskCh chan *task.RequestTask
     	// task map taskId-> RequestTaskWork
     	taskMap map[int64]*task.RequestTaskWork
     
     	taskMapMu *sync.RWMutex // 读写锁
     }
     ```

     - ##### 发送节点自身信息

       定时向特定redis channel发送自身节点信息；

     - ##### 尝试成为主节点TryBeMaster

       订阅redis 主节点选举结果channel，从中读取是否自己被选举成主节点。该逻辑由哨兵节点选举推送；

     - ##### 接收任务receiveTask

       订阅自己节点的redis channel，从该管道读取任务，推送至taskCh

     - ##### 处理任务

       从taskCh读取任务，判断该任务是新增任务，还是已有任务更改QPS或者停止：

       ​         新增任务：创建一个新的任务work，n qps创建n个goroutine；

       ```go
       // work node节点实际运行一个新任务结构体
       type RequestTaskWork struct {
       	RequestTask
       	Ctx context.Context
       	// qps改变信号通道,用于新增或释放goroutine
       	QPSReduceCh chan struct{}
       	ResultCh    chan *response.Result
       	stop        func()
       }
       
       // 开始一个任务
       func (rtw *RequestTaskWork) Start() {
       	// 处理任务结果
       	go rtw.UploadResults(rtw.Ctx)
       
       	for i := 0; i < rtw.TargetQPS; i++ {
       		go rtw.SendReq(rtw.Ctx)
       	}
       
       }
       ```

       ​         已有任务更改QPS ：新增QPS则创建额外的goroutine执行，减少QPS则通过 RequestTaskWork.QPSReduceCh 传送信号，工作goroutine接收后退出部分。

   4. #### 任务分配策略

      当前任务的默认分配策略，根据节点的Cpu使用百分比/ 内存使用百分比/ 内存容量划分优先级
  
      
  
- ### 设计参考   

  https://time.geekbang.org/column/article/270474

  http://go-database-sql.org/index.html

  https://github.com/go-sql-driver/mysql

  https://github.com/link1st/go-stress-testing

