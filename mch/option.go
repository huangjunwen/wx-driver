package mch

import (
	"github.com/huangjunwen/wxdriver"
)

var (
	// DefaultOptions 为 mch 模块的默认 Options
	DefaultOptions Options
)

// Options 包含调用微信支付接口时的选项，NOTE: 某些选项未必对所有接口都有意义
type Options struct {
	// HTTPClient 为调用微信支付接口时使用的 http 客户端，若空则使用默认的；
	wxdriver.HTTPClient

	// APIVersion 指定接口版本号，某些接口有多个版本，默认为 APIVersionDefault
	APIVersion

	// SignType 指定签名类型，默认为 MD5
	SignType
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

// UseAPIVersion 设置接口版本号，NOTE：不是所有接口都有版本
func UseAPIVersion(ver APIVersion) Option {
	return func(opts *Options) error {
		opts.APIVersion = ver
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
