package task

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/lizaiganshenmo/GoStress/cmd/stress/task/response"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/lizaiganshenmo/GoStress/cmd/stress/task/verify"
	"github.com/lizaiganshenmo/GoStress/library/errno"
	"golang.org/x/net/http2"
)

// 发送http请求
func (rt *RequestTask) SendHttpReq() *response.Result {
	var res response.Result
	method := rt.Method
	url := rt.URL
	reqBody := rt.GetBody()
	timeout := rt.Timeout
	headers := rt.Headers

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return &res
	}

	// 在req中设置Host，解决在header中设置Host不生效问题
	if _, ok := headers["Host"]; ok {
		req.Host = headers["Host"]
	}
	// 设置默认为utf-8编码
	if _, ok := headers["Content-Type"]; !ok {
		if headers == nil {
			headers = make(map[string]string)
		}
		headers["Content-Type"] = "application/x-www-form-urlencoded; charset=utf-8"
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	var client *http.Client
	if rt.Keepalive {
		client = NewLongClient(rt.TaskID, rt)
	} else {
		req.Close = true
		tr := &http.Transport{}
		if rt.UseHTTP2 {
			// 使用真实证书 验证证书 模拟真实请求
			tr = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
			}
			if err = http2.ConfigureTransport(tr); err != nil {
				res.ErrCode = errno.RequestErr
				return &res
			}
		} else {
			// 跳过证书验证
			tr = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}

		client = &http.Client{
			Transport: tr,
			Timeout:   timeout,
		}
	}

	startTime := time.Now()
	resp, err := client.Do(req)

	res.TimeConsuming = time.Since(startTime).Milliseconds()
	if err != nil {
		hlog.Warnf("请求失败. err:%+v", err)
		res.ErrCode = errno.RequestErr
		return &res
	}

	// 校验请求结果
	code, isSuccessed := verify.VerifyForHTTP(rt.Verify, resp)

	res.ErrCode = code
	res.IsSucceed = isSuccessed
	res.ReqTime = startTime
	return &res
}
