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

// configuration 是默认 Configuration 实现
type configuration struct {
	appID    string
	payMchID string
	payKey   string
}

// NewConfiguration 构造一个 Configuration
func NewConfiguration(appID, payMchID, payKey string) Configuration {
	return &configuration{
		appID:    appID,
		payMchID: payMchID,
		payKey:   payKey,
	}
}

func (c *configuration) WechatAppID() string {
	return c.appID
}

func (c *configuration) WechatPayMchID() string {
	return c.payMchID
}

func (c *configuration) WechatPayKey() string {
	return c.payKey
}
