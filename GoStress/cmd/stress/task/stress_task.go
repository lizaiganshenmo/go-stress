package task

import (
	"github.com/lizaiganshenmo/GoStress/cmd/stress/task/response"

	"context"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/lizaiganshenmo/GoStress/library"
	"github.com/lizaiganshenmo/GoStress/resource"
	"github.com/lizaiganshenmo/GoStress/resource/db"
)

type RequestTask struct {
	TaskID     int64             // task id
	TargetQPS  int               // 目标QPS
	URL        string            // URL
	Protocol   string            // http/webSocket/tcp
	Method     string            // 方法 GET/POST/PUT
	Headers    map[string]string // Headers
	Body       string            // body
	Verify     string            // 验证的方法
	Timeout    time.Duration     // 请求超时时间
	UseHTTP2   bool              // 是否使用http2.0
	Keepalive  bool              // 是否开启长连接
	TaskStatus int               // 任务状态-执行中、更改qps、待停止
}

// GetBody 获取请求数据
func (r *RequestTask) GetBody() (body io.Reader) {
	return strings.NewReader(r.Body)
}

// work node节点实际运行一个新任务结构体
type RequestTaskWork struct {
	RequestTask
	Ctx context.Context
	// qps改变信号通道,用于新增或释放goroutine
	QPSReduceCh chan struct{}
	ResultCh    chan *response.Result
	stop        func()
}

// new RequestTaskWork
func NewRequestTaskWork(ctx context.Context, cancelFunc func(), task *RequestTask) *RequestTaskWork {
	return &RequestTaskWork{
		Ctx:         ctx,
		RequestTask: *task,
		QPSReduceCh: make(chan struct{}, 1000),
		ResultCh:    make(chan *response.Result, 1000),
		stop:        cancelFunc,
	}

}

// 开始一个任务
func (rtw *RequestTaskWork) Start() {
	// 处理任务结果
	go rtw.UploadResults(rtw.Ctx)

	for i := 0; i < rtw.TargetQPS; i++ {
		go rtw.SendReq(rtw.Ctx)
	}

}

// 任务停止
func (rtw *RequestTaskWork) Stop() {
	rtw.stop()
}

// 任务QPS更改
func (rtw *RequestTaskWork) ChangeQPS(newQPS int) {
	t := newQPS - rtw.TargetQPS
	rtw.TargetQPS = newQPS
	// qps增加
	if t > 0 {
		for i := 0; i < t; i++ {
			go rtw.SendReq(rtw.Ctx)
		}
	} else {
		t = 0 - t
		for i := 0; i < t; i++ {
			rtw.QPSReduceCh <- struct{}{}
		}

	}

}

// 发送请求,每秒一次
func (rtw *RequestTaskWork) SendReq(ctx context.Context) {
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-rtw.QPSReduceCh:
			// 接收到QPS降低信号，退出
			return
		case <-ticker.C:
			rtw.SendOnce(ctx)
		}
	}

}

// 发送一次请求
func (rtw *RequestTaskWork) SendOnce(ctx context.Context) {
	switch rtw.Protocol {
	case library.ProtocolHTTP:
		rtw.ResultCh <- rtw.SendHttpReq()
	case library.ProtocolWebSocket:
	case library.ProtocolTCP:

	}

}

// 批量上传结果  ->此处上传至influxDB
func (rtw *RequestTaskWork) UploadResults(ctx context.Context) {
	defaultSize := 100
	resArr := make([]*response.Result, 0, defaultSize)
	for {
		select {
		case <-ctx.Done():
			return
		case res := <-rtw.ResultCh:
			resArr = append(resArr, res)

			if len(resArr) == defaultSize {
				err := rtw.batchUploadRes(ctx, resArr)
				if err != nil {
					hlog.Warnf("rtw.batchUploadRes fail.err:%+v", err)
				}

				resArr = make([]*response.Result, 0, defaultSize)
			}

		}
	}

}

// 批量上传res
func (rtw *RequestTaskWork) batchUploadRes(ctx context.Context, resArr []*response.Result) error {
	taskIDStr := strconv.FormatInt(rtw.TaskID, 10)
	tags := map[string]string{"taskID": taskIDStr}

	writeAPI := (*resource.InfluxDBCli).WriteAPIBlocking(library.InfluxDBOrg, library.InfluxDBBucket)
	points := make([]*write.Point, 0, len(resArr))
	for _, v := range resArr {
		t := v
		fields := map[string]interface{}{
			"timeConsuming": t.TimeConsuming,
			"isSucceed":     t.IsSucceed,
			"errCode":       t.ErrCode,
		}
		pt := write.NewPoint(library.InfluxDBMeasurment, tags, fields, t.ReqTime)
		points = append(points, pt)

		// 	bp.AddPoint(pt)
		// }

		// writeAPI.EnableBatching()

	}

	err := writeAPI.WritePoint(ctx, points...)

	return err

}

// trans *db.StressTask -> *RequestTask
func FormatRequestTask(task *db.StressTask) *RequestTask {
	headers := make(map[string]string)
	err := sonic.UnmarshalString(task.Headers, &headers)
	if err != nil {
		hlog.Warnf("invalid headers. task: %+v, err:%+v", task, err)
		return nil
	}

	return &RequestTask{
		TaskID:     task.TaskID,
		TargetQPS:  task.TargetQPS,
		URL:        task.URL,
		Protocol:   task.Protocol,
		Method:     task.Method,
		Headers:    headers,
		Body:       task.Body,
		Verify:     task.Verify,
		Timeout:    time.Duration(task.Timeout * int(time.Second)),
		UseHTTP2:   task.UseHTTP2,
		Keepalive:  task.KeepAlive,
		TaskStatus: task.Status,
	}
}
