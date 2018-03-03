package mch

// Version 代表 API 版本号
type Version struct{ v string }

// TradeType 表示交易方式
type TradeType struct{ v string }

// TradeState 表示交易状态
type TradeState struct{ v string }

// SignType 代表签名类型
type SignType struct{ v string }

const (
	datetimeLayout = "20060102150405"
)

const (
	// URLBaseDefault 默认接入点
	URLBaseDefault = "https://api.mch.weixin.qq.com"
	// URLBaseHK 建议东南亚接入点
	URLBaseHK = "https://apihk.mch.weixin.qq.com"
	// URLBaseUS 建议其它地区接入点
	URLBaseUS = "https://apius.mch.weixin.qq.com"
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

// ParseTradeType parse 交易类型字符串
func ParseTradeType(v string) TradeType {
	switch v {
	case "JSAPI", "NATIVE", "APP", "MWEB":
		return TradeType{v}
	default:
		return TradeType{}
	}
}

// String 实现 Stringer 接口
func (tt TradeType) String() string {
	return tt.v
}

// IsValid 当该值有效(非空)时返回 true
func (tt TradeType) IsValid() bool {
	return tt.v != ""
}

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

// ParseTradeState parse 交易状态字符串
func ParseTradeState(v string) TradeState {
	switch v {
	case "NOTPAY", "CLOSED", "PAYERROR", "SUCCESS", "REFUND", "USERPAYING", "REVOKED":
		return TradeState{v}
	default:
		return TradeState{}
	}
}

// String 实现 Stringer 接口
func (ts TradeState) String() string {
	return ts.v
}

// IsValid 当该值有效(非空)时返回 true
func (ts TradeState) IsValid() bool {
	return ts.v != ""
}

var (
	SignTypeInvalid    = SignType{""}
	SignTypeMD5        = SignType{"MD5"}
	SignTypeHMACSHA256 = SignType{"HMAC-SHA256"}
)

// String 实现 Stringer 接口
func (st SignType) String() string {
	return st.v
}

// IsValid 当该值有效(非空)时返回 true
func (st SignType) IsValid() bool {
	return st.v != ""
}
