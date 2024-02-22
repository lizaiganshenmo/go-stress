package verify

import (
	"net/http"
	"sync"

	"github.com/lizaiganshenmo/GoStress/library/errno"
)

const (
	HttpStatusVerifyName = "statusCode"
	HttpJsonVerifyName   = "httpJson"
)

var (
	// verifyMapHTTP http 校验函数
	verifyMapHTTP = make(map[string]HTTPVerify)
	// verifyMapHTTPMutex http 并发锁
	verifyMapHTTPMu sync.RWMutex
)

func init() {
	RegisterVerifyHTTP(HttpStatusVerifyName, VerifyStatusCode)
	RegisterVerifyHTTP(HttpJsonVerifyName, VerifyJson)
}

// RegisterVerifyHTTP 注册 http 校验函数
func RegisterVerifyHTTP(verify string, verifyFunc HTTPVerify) {
	verifyMapHTTPMu.Lock()
	verifyMapHTTP[verify] = verifyFunc
	verifyMapHTTPMu.Unlock()
}

// 校验http
func VerifyForHTTP(verifyName string, resp *http.Response) (code int, isSucceed bool) {
	verifyMapHTTPMu.RLock()
	verify, ok := verifyMapHTTP[verifyName]
	verifyMapHTTPMu.RUnlock()

	if !ok {
		code = errno.VerifyNotExist
		return
	}

	return verify(resp)

}
