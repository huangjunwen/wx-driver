package wxdriver

import (
	"net/http"
)

// HTTPClient 是对 http.Client 的一个泛化；
// 应用可以对 http.Client 进一步封装，例如添加请求/响应的日志记录等
type HTTPClient interface {
	// Do 发送请求，等待响应或错误；实现的行为应当与 http.Client.Do 一致
	Do(req *http.Request) (*http.Response, error)
}
