package verify

import (
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/lizaiganshenmo/GoStress/library/errno"
)

type ResponseJSON struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
type HTTPVerify func(resp *http.Response) (code int, isSucceed bool)

// 校验http返回状态码
func VerifyStatusCode(resp *http.Response) (code int, isSucceed bool) {
	code = resp.StatusCode
	if resp.StatusCode == errno.SuccessStatusCode {
		isSucceed = true
	}

	return
}

// 校验返回结果
func VerifyJson(resp *http.Response) (code int, isSucceed bool) {
	code = resp.StatusCode

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		code = errno.ParseError
		return
	}

	var res ResponseJSON
	err = sonic.Unmarshal(respBody, &res)
	if err != nil {
		code = errno.ParseError
		return
	}

	code = res.Code
	if res.Code == errno.SuccessCode {
		isSucceed = true
	}

	return
}
