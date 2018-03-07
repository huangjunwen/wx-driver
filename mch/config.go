package mch

// Config 包含微信支付接口所必须的配置信息
type Config interface {
	// WechatAppID 返回微信应用（公众号/小程序...） app id
	WechatAppID() string

	// WechatMchID 返回微信支付商户号
	WechatMchID() string

	// WechatMchKey 返回微信支付密钥
	WechatMchKey() string
}

// ConfigSelector 用于选择配置
type ConfigSelector interface {
	// Select 通过 appID 和 MchID 查找对应配置，若找不到应该返回 nil
	Select(appID, MchID string) Config
}

// defaultConfig 是默认 Configuration 实现
type defaultConfig struct {
	appID  string
	mchID  string
	mchKey string
}

// singleConfigSelector 包含单个 Configuration
type singleConfigSelector struct {
	config Config
}

// NewConfig 构造一个 Configuration
func NewConfig(appID, mchID, mchKey string) Config {
	return &defaultConfig{
		appID:  appID,
		mchID:  mchID,
		mchKey: mchKey,
	}
}

func (c *defaultConfig) WechatAppID() string {
	return c.appID
}

func (c *defaultConfig) WechatMchID() string {
	return c.mchID
}

func (c *defaultConfig) WechatMchKey() string {
	return c.mchKey
}

// NewSingleConfigSelector 返回一个单配置 selector
func NewSingleConfigSelector(config Config) ConfigSelector {
	return singleConfigSelector{config: config}
}

func (s singleConfigSelector) Select(appID, mchID string) Config {
	if s.config.WechatAppID() == appID && s.config.WechatMchID() == mchID {
		return s.config
	}
	return nil
}
