package mch

import (
	"github.com/huangjunwen/wxdriver"
	"net/http"
	"net/url"
)

var (
	// DefaultOptions 为 mch 模块的默认设置，可修改它影响该模块下的默认行为，
	DefaultOptions = &Options{
		urlBase:  URLBaseDefault,
		signType: SignTypeMD5,
	}
)

// Options 包含调用微信支付接口时的选项，NOTE: 某些选项未必对所有接口都有意义
type Options struct {
	// http 客户端
	client wxdriver.HTTPClient

	// 地址前缀
	urlBase string

	// 签名类型
	signType SignType
}

// Option 代表调用微信支付接口时的单个选项
type Option func(*Options) error

func newOptions(opts []Option) (*Options, error) {
	if len(opts) == 0 {
		return nil, nil
	}

	options := &Options{}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}
	return options, nil
}

func mustOptions(opts []Option) *Options {
	options, err := newOptions(opts)
	if err != nil {
		panic(err)
	}
	return options
}

// NewOptions 创建一个 Options
//
// NOTE: 若 len(opts) == 0，返回 (*Options)(nil) 也是有效的
func NewOptions(opts ...Option) (*Options, error) {
	return newOptions(opts)
}

// MustOptions 是 must 版 NewOptions
func MustOptions(opts ...Option) *Options {
	return mustOptions(opts)
}

// Client 返回 HTTPClient，依次：options.client > DefaultOptions.client > wxdriver.DefaultHTTPClient > http.DefaultClient
//
// NOTE: 即使 options 为 nil 指针该方法仍能有效返回
func (options *Options) Client() wxdriver.HTTPClient {
	if options != nil && options.client != nil {
		return options.client
	}
	if DefaultOptions != nil && DefaultOptions.client != nil {
		return DefaultOptions.client
	}
	if wxdriver.DefaultHTTPClient != nil {
		return wxdriver.DefaultHTTPClient
	}
	return http.DefaultClient
}

// URLBase 返回 API 地址前缀，依次：options.urlBase > DefaultOptions.urlBase > "https://api.mch.weixin.qq.com"
//
// NOTE: 即使 options 为 nil 指针该方法仍能有效返回
func (options *Options) URLBase() string {
	if options != nil && options.urlBase != "" {
		return options.urlBase
	}
	if DefaultOptions != nil && DefaultOptions.urlBase != "" {
		return DefaultOptions.urlBase
	}
	return URLBaseDefault
}

// SignType 返回签名方式，依次：options.signType > DefaultOptions.signType > MD5
//
// NOTE: 即使 options 为 nil 指针该方法仍能有效返回
func (options *Options) SignType() SignType {
	if options != nil && options.signType.IsValid() {
		return options.signType
	}
	if DefaultOptions != nil && DefaultOptions.signType.IsValid() {
		return DefaultOptions.signType
	}
	return SignTypeMD5
}

// UseClient 设置 HTTPClient
func UseClient(client wxdriver.HTTPClient) Option {
	return func(options *Options) error {
		options.client = client
		return nil
	}
}

// UseURLBase 设置 API 地址前缀，默认为 "https://api.mch.weixin.qq.com"
func UseURLBase(urlBase string) Option {
	return func(options *Options) error {
		u, err := url.Parse(urlBase)
		if err != nil {
			return err
		}
		// 只取 scheme 和 host 部分
		options.urlBase = (&url.URL{
			Scheme: u.Scheme,
			Host:   u.Host,
		}).String()
		return nil
	}
}

// UseSignType 设置签名类型
func UseSignType(signType SignType) Option {
	return func(options *Options) error {
		options.signType = signType
		return nil
	}
}
