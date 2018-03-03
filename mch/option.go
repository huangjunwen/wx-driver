package mch

import (
	"github.com/huangjunwen/wxdriver"
	"net/url"
)

var (
	// DefaultOptions 为 mch 模块的默认 Options
	DefaultOptions Options
)

// Options 包含调用微信支付接口时的选项，NOTE: 某些选项未必对所有接口都有意义
type Options struct {
	// HTTPClient 为调用微信支付接口时使用的 http 客户端，若空则使用默认的；
	HTTPClient wxdriver.HTTPClient

	// URLBase 是接口 scheme + host，如空则使用默认 "https://api.mch.weixin.qq.com"
	URLBase string

	// SignType 指定签名类型，如空则使用默认 MD5
	SignType SignType
}

// Option 代表调用微信支付接口时的单个选项
type Option func(*Options) error

// UseClient 设置 HTTPClient
func UseClient(client wxdriver.HTTPClient) Option {
	return func(opts *Options) error {
		opts.HTTPClient = client
		return nil
	}
}

// UseURLBase 设置 URLBase，默认为 "https://api.mch.weixin.qq.com"
func UseURLBase(urlBase string) Option {
	return func(opts *Options) error {
		u, err := url.Parse(urlBase)
		if err != nil {
			return err
		}
		// 只取 scheme 和 host 部分
		opts.URLBase = (&url.URL{
			Scheme: u.Scheme,
			Host:   u.Host,
		}).String()
		return nil
	}
}

// UseSignType 设置签名类型
func UseSignType(signType SignType) Option {
	return func(opts *Options) error {
		opts.SignType = signType
		return nil
	}
}

// NewOptions 组装 Options
func NewOptions(options ...Option) (*Options, error) {
	// 以默认选项为蓝本
	ret := DefaultOptions
	for _, option := range options {
		if err := option(&ret); err != nil {
			return nil, err
		}
	}
	return &ret, nil
}
