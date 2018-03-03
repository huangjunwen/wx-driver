package mch

const (
	// URLBaseDefault 默认接入点
	URLBaseDefault = "https://api.mch.weixin.qq.com"
	// URLBaseHK 建议东南亚接入点
	URLBaseHK = "https://apihk.mch.weixin.qq.com"
	// URLBaseUS 建议其它地区接入点
	URLBaseUS = "https://apius.mch.weixin.qq.com"
)

const (
	DatetimeLayout = "20060102150405"
)

var (
	VersionDefault = Version{""}
	Version1       = Version{"1.0"}
)

var (
	// TradeTypeInvalid 表示无效交易类型
	TradeTypeInvalid = TradeType{""}
	// TradeTypeJSAPI 表示公众号/小程序交易类型
	TradeTypeJSAPI = TradeType{"JSAPI"}
	// TradeTypeNATIVE 表示扫码交易类型
	TradeTypeNATIVE = TradeType{"NATIVE"}
	// TradeTypeAPP 表示 app 交易类型
	TradeTypeAPP = TradeType{"APP"}
	// TradeTypeMWEB 表示 H5 交易类型
	TradeTypeMWEB = TradeType{"MWEB"}
)

var (
	// TradeStateInvalid 表示无效交易状态
	TradeStateInvalid = TradeState{""}
	// TradeStateNOTPAY 表示未支付
	TradeStateNOTPAY = TradeState{"NOTPAY"}
	// TradeStateCLOSED 表示已关闭；NOTPAY 关单 -> CLOSED
	TradeStateCLOSED = TradeState{"CLOSED"}
	// TradeStatePAYERROR 表示支付错误（如银行返回失败）；NOTPAY -> PAYERROR
	TradeStatePAYERROR = TradeState{"PAYERROR"}
	// TradeStateSUCCESS 表示支付成功；NOTPAY -> SUCCESS
	TradeStateSUCCESS = TradeState{"SUCCESS"}
	// TradeStateREFUND 表示进入退款；SUCCESS 发起退款 -> REFUND
	TradeStateREFUND = TradeState{"REFUND"}
	// TradeStateUSERPAYING 表示用户支付中？
	TradeStateUSERPAYING = TradeState{"USERPAYING"}
	// TradeStateREVOKED 表示已撤销（刷卡支付）
	TradeStateREVOKED = TradeState{"REVOKED"}
)

var (
	SignTypeInvalid    = SignType{""}
	SignTypeMD5        = SignType{"MD5"}
	SignTypeHMACSHA256 = SignType{"HMAC-SHA256"}
)
