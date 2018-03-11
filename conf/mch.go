package conf

// MchConfig 包含微信支付接口所必须的配置信息
type MchConfig interface {
	// WechatAppID 返回微信应用（公众号/小程序...） app id
	WechatAppID() string

	// WechatMchID 返回微信支付商户号
	WechatMchID() string

	// WechatMchKey 返回微信支付密钥
	WechatMchKey() string
}

// MchConfigSelector 用于选择微信支付配置
type MchConfigSelector interface {
	// SelectMch 通过 appID 和 mchID 查找对应配置，若找不到应该返回 nil
	SelectMch(appID, mchID string) MchConfig
}
