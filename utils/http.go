package utils

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	// DefaultHTTPClient 为 wxdriver 默认 HTTPClient，30 秒超时
	DefaultHTTPClient HTTPClient = &http.Client{
		Timeout: 30 * time.Second,
	}
)

// HTTPClient 是对 http.Client 的一个泛化；
// 应用可以对 http.Client 进一步封装，例如添加请求/响应的日志记录等
type HTTPClient interface {
	// Do 发送请求，等待响应或错误；实现的行为应当与 http.Client.Do 一致
	Do(req *http.Request) (*http.Response, error)
}

// ReadAndReplaceRequestBody 读取 Request 全部 body 并将 body 替换成 bytes.Buffer
func ReadAndReplaceRequestBody(req *http.Request) (reqBody []byte, err error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return body, nil
}

// ReadAndReplaceResponseBody 读取 Response 全部 body 并将 body 替换成 bytes.Buffer
func ReadAndReplaceResponseBody(resp *http.Response) (respBody []byte, err error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return body, nil
}
