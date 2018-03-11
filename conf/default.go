package conf

const (
	// WechatAppIDName 下存的是微信应用（公众号/小程序...） app id
	WechatAppIDName = "WechatAppID"
	// WechatMchIDName 下存的是微信支付商户号
	WechatMchIDName = "WechatMchID"
	// WechatMchKeyName 下存的是微信支付密钥
	WechatMchKeyName = "WechatMchKey"
)

// DefaultConfig 用 map[string]string 存配置
type DefaultConfig map[string]string

// WechatAppID 返回微信应用（公众号/小程序...） app id
func (config DefaultConfig) WechatAppID() string {
	return config[WechatAppIDName]
}

// SetWechatAppID 设置微信应用（公众号/小程序...） app id
func (config DefaultConfig) SetWechatAppID(v string) {
	config[WechatAppIDName] = v
}

// WechatMchID 返回微信支付商户号
func (config DefaultConfig) WechatMchID() string {
	return config[WechatMchIDName]
}

// SetWechatMchID 设置微信支付商户号
func (config DefaultConfig) SetWechatMchID(v string) {
	config[WechatMchIDName] = v
}

// WechatMchKey 返回微信支付密钥
func (config DefaultConfig) WechatMchKey() string {
	return config[WechatMchKeyName]
}

// SetWechatMchKey 设置微信支付密钥
func (config DefaultConfig) SetWechatMchKey(v string) {
	config[WechatMchKeyName] = v
}

// SelectMch 实现 MchConfigSelector 接口
func (config DefaultConfig) SelectMch(appID, mchID string) MchConfig {
	if appID == "" || mchID == "" {
		return nil
	}
	if config.WechatAppID() == appID && config.WechatMchID() == mchID {
		return config
	}
	return nil
}
