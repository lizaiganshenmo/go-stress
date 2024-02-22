package utils

import (
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/lizaiganshenmo/GoStress/library"
)

// CURL curl参数解析
type CURL struct {
	Data map[string][]string
}

func NewCURL() *CURL {
	return &CURL{
		Data: map[string][]string{},
	}
}

// getDataValue 获取数据
func (c *CURL) getDataValue(keys []string) []string {
	var (
		value = make([]string, 0)
	)
	for _, key := range keys {
		var (
			ok bool
		)
		value, ok = c.Data[key]
		if ok {
			break
		}
	}
	return value
}

// Parse 从文件文本中解析curl
func Parse(dataBytes []byte) *CURL {
	curl := NewCURL()

	data := string(dataBytes)
	for len(data) > 0 {
		if strings.HasPrefix(data, "curl") {
			data = data[5:]
		}
		data = strings.TrimSpace(data)
		var (
			key   string
			value string
		)
		index := strings.Index(data, " ")
		if index <= 0 {
			break
		}
		key = strings.TrimSpace(data[:index])
		data = data[index+1:]
		data = strings.TrimSpace(data)
		// url
		if !strings.HasPrefix(key, "-") {
			key = strings.Trim(key, "'")
			curl.Data["curl"] = []string{key}
			// 去除首尾空格
			data = strings.TrimFunc(data, func(r rune) bool {
				if r == ' ' || r == '\\' || r == '\n' {
					return true
				}
				return false
			})
			continue
		}
		if strings.HasPrefix(data, "-") {
			continue
		}
		var (
			endSymbol = " "
		)
		if strings.HasPrefix(data, "'") {
			endSymbol = "'"
			data = data[1:]
		}
		index = strings.Index(data, endSymbol)
		if index <= -1 {
			index = len(data)
			// break
		}
		value = data[:index]
		if len(data) >= index+1 {
			data = data[index+1:]
		} else {
			data = ""
		}
		// 去除首尾空格
		data = strings.TrimFunc(data, func(r rune) bool {
			if r == ' ' || r == '\\' || r == '\n' {
				return true
			}
			return false
		})
		if key == "" {
			continue
		}
		curl.Data[key] = append(curl.Data[key], value)
	}
	return curl
}

// String string
func (c *CURL) String() (url string) {
	curlByte, _ := sonic.Marshal(c)
	return string(curlByte)
}

// GetURL 获取url
func (c *CURL) GetURL() (url string) {
	keys := []string{"curl", "--url", "--location"}
	value := c.getDataValue(keys)
	if len(value) <= 0 {
		return
	}
	url = value[0]
	return
}

// GetMethod 获取 请求方式
func (c *CURL) GetMethod() (method string) {
	keys := []string{"-X", "--request"}
	value := c.getDataValue(keys)
	if len(value) <= 0 {
		return c.defaultMethod()
	}
	method = strings.ToUpper(value[0])
	if library.InArrayStr(method, []string{"GET", "POST", "PUT", "DELETE"}) {
		return method
	}
	return c.defaultMethod()
}

// defaultMethod 获取默认方法
func (c *CURL) defaultMethod() (method string) {
	method = "GET"
	body := c.GetBody()
	if len(body) > 0 {
		return "POST"
	}
	return
}

// GetHeaders 获取请求头
func (c *CURL) GetHeaders() (headers map[string]string) {
	headers = make(map[string]string, 0)
	keys := []string{"-H", "--header"}
	value := c.getDataValue(keys)
	for _, v := range value {
		getHeaderValue(v, headers)
	}
	return
}

// GetHeadersStr 获取请求头string
func (c *CURL) GetHeadersStr() string {
	headers := c.GetHeaders()
	bytes, _ := sonic.Marshal(&headers)
	return string(bytes)
}

// GetBody 获取body
func (c *CURL) GetBody() (body string) {
	keys := []string{"--data", "-d", "--data-urlencode", "--data-raw", "--data-binary"}
	value := c.getDataValue(keys)
	if len(value) <= 0 {
		body = c.getPostForm()
		return
	}
	body = value[0]
	return
}

// getPostForm get post form
func (c *CURL) getPostForm() (body string) {
	keys := []string{"--form", "-F", "--form-string"}
	value := c.getDataValue(keys)
	if len(value) <= 0 {
		return
	}
	body = strings.Join(value, "&")
	return
}

func getHeaderValue(v string, headers map[string]string) {
	index := strings.Index(v, ":")
	if index < 0 {
		return
	}
	vIndex := index + 1
	if len(v) >= vIndex {
		value := strings.TrimPrefix(v[vIndex:], " ")
		if _, ok := headers[v[:index]]; ok {
			headers[v[:index]] = fmt.Sprintf("%s; %s", headers[v[:index]], value)
		} else {
			headers[v[:index]] = value
		}
	}
}
