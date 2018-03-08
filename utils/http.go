package utils

import (
	"crypto/tls"
	"crypto/x509"
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

// NewHTTPSClient 以客户端 ssl 证书和密钥创建一个 https client，caPEMBlock 如果非空，
// 则以之为 ca，否则使用 host 本身的 ca
func NewHTTPSClient(certPEMBlock, keyPEMBlock, caPEMBlock []byte) (*http.Client, error) {
	// 创建证书
	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return nil, err
	}

	// 如果有 ca 提供，否则使用 host 本身的 ca
	var caCertPool *x509.CertPool
	if len(caPEMBlock) != 0 {
		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caPEMBlock)
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
				RootCAs:      caCertPool,
			},
		},
	}, nil

}
