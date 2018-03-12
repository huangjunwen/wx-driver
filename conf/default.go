package conf

// DefaultConfig 是默认配置，它实现多个子模块的配置接口
type DefaultConfig struct {
	// AppID 微信应用（公众号/小程序...） app id
	AppID string
	// MchID 微信支付商户号
	MchID string
	// MchKey 微信支付密钥
	MchKey string
}

var (
	_ MchConfig         = (*DefaultConfig)(nil)
	_ MchConfigSelector = (*DefaultConfig)(nil)
)

// WechatAppID 返回微信应用（公众号/小程序...） app id
func (config *DefaultConfig) WechatAppID() string {
	return config.AppID
}

// WechatMchID 返回微信支付商户号
func (config *DefaultConfig) WechatMchID() string {
	return config.MchID
}

// WechatMchKey 返回微信支付密钥
func (config *DefaultConfig) WechatMchKey() string {
	return config.MchKey
}

// SelectMch 实现 MchConfigSelector 接口
func (config *DefaultConfig) SelectMch(appID, mchID string) MchConfig {
	if appID == "" || mchID == "" {
		return nil
	}
	if config.WechatAppID() == appID && config.WechatMchID() == mchID {
		return config
	}
	return nil
}
