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

// ConfigurationSelector 用于选择配置
type ConfigurationSelector interface {
	// Select 通过 appID 和 MchID 查找对应配置，若找不到应该返回 nil
	Select(appID, MchID string) Configuration
}

// configuration 是默认 Configuration 实现
type configuration struct {
	appID    string
	payMchID string
	payKey   string
}

// singleConfigurationSelector 包含单个 Configuration
type singleConfigurationSelector struct {
	config Configuration
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

// NewSingleConfigurationSelector 返回一个单配置 selector
func NewSingleConfigurationSelector(config Configuration) ConfigurationSelector {
	return singleConfigurationSelector{config: config}
}

func (s singleConfigurationSelector) Select(appID, mchID string) Configuration {
	if s.config.WechatAppID() == appID && s.config.WechatPayMchID() == mchID {
		return s.config
	}
	return nil
}
