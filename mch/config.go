package mch

// Configuration 包含微信支付接口所必须的配置信息
type Configuration interface {
	// WechatAppID 返回微信应用（公众号/小程序...） app id
	WechatAppID() string

	// WechatPayMchID 返回微信支付商户号
	WechatPayMchID() string

	// WechatPayKey 返回微信支付密钥
	WechatPayKey() string
}
