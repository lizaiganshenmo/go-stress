package response

import "time"

type Result struct {
	ReqTime       time.Time // 请求时间 毫秒
	TimeConsuming int64     // 请求耗时
	IsSucceed     bool      // 是否请求成功
	ErrCode       int       // 错误码
}
