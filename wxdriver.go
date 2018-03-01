package wxdriver

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

var (
	DefaultHTTPClient HTTPClient
)

// HTTPClient 是对 http.Client 的一个泛化；
// 应用可以对 http.Client 进一步封装，例如添加请求/响应的日志记录等
type HTTPClient interface {
	// Do 发送请求，等待响应或错误；实现的行为应当与 http.Client.Do 一致
	Do(req *http.Request) (*http.Response, error)
}

// NonceStr 生成一个长度为 2*n 的随机字符串
func NonceStr(n int) string {
	buf := make([]byte, n)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}
