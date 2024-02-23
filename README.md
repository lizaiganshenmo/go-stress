# go-stress
### 项目介绍

​            分布式压测系统-golang版。

​            项目旨在提供分布式压测方式，通过主从节点，高可用支持压测任务。golang版本实现，以go高并发的优良特性，通过一个协程对标模拟一个用户，进行压测。

- [x] HTTP/1 压测

- [x] HTTP/2 压测

- [x] 单机压测

- [x] 分布式压测

- [x] HTTP服务操作压测任务

- [x] 可视化展示

- [x] 阶梯式发压

- [x] 高可用性（还行）

- [ ] HTTP/3 压测

- [ ] Websocket 压测

### 如何使用

1. #### 基础环境

   一台机器或多台、MySQL、Redis、InfluxDB、ElasticSearch(上传日志，可有可无)、grafana；

2. #### 创建任务、执行、停止

   通过http服务，调相关接口创建执行即可。
   https://github.com/lizaiganshenmo/go-stress/blob/main/stress.postman_collection.json

4. #### Demo展示

   该图是grafana接入InfluxDB数据源的展示，任务执行中支持更改QPS，可以阶梯式压测。

   ![image-20240223204737525](https://github.com/lizaiganshenmo/pictures/blob/main/%E5%8E%8B%E6%B5%8Bdemo.jpg)

   

   

   

  

