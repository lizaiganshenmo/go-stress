package task

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

var (
	mutex   sync.RWMutex
	clients = make(map[int64]*http.Client, 0)
)

// NewClient new
func NewLongClient(i int64, request *RequestTask) *http.Client {
	client := getLongClient(i)
	if client != nil {
		return client
	}
	return setClient(i, request)
}

func getLongClient(i int64) *http.Client {
	mutex.RLock()
	defer mutex.RUnlock()
	return clients[i]
}

func setClient(i int64, request *RequestTask) *http.Client {
	mutex.Lock()
	defer mutex.Unlock()
	client := createLangHttpClient(request)
	clients[i] = client
	return client
}

// createLangHttpClient 初始化长连接客户端参数
func createLangHttpClient(request *RequestTask) *http.Client {
	tr := &http.Transport{}
	if request.UseHTTP2 {
		// 使用真实证书 验证证书 模拟真实请求
		tr = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        0,                // 最大连接数,默认0无穷大
			MaxIdleConnsPerHost: 100,              // 对每个host的最大连接数量(MaxIdleConnsPerHost<=MaxIdleConns)
			IdleConnTimeout:     90 * time.Second, // 多长时间未使用自动关闭连接
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
		}
		_ = http2.ConfigureTransport(tr)
	} else {
		// 跳过证书验证
		tr = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        0,                // 最大连接数,默认0无穷大
			MaxIdleConnsPerHost: 100,              // 对每个host的最大连接数量(MaxIdleConnsPerHost<=MaxIdleConns)
			IdleConnTimeout:     90 * time.Second, // 多长时间未使用自动关闭连接
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		}
	}
	return &http.Client{
		Transport: tr,
	}
}
