package mch

// TradeType 表示交易方式
type TradeType struct{ v string }

// TradeState 表示交易状态
type TradeState struct{ v string }

// RefundStatus 表示退款状态
type RefundStatus struct{ v string }

// SignType 代表签名类型
type SignType struct{ v string }

// Version 代表 API 版本号，TODO: 接口加上 Version 逻辑
type Version struct{ v string }

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

// ParseRefundStatus parse 退款状态
func ParseRefundStatus(v string) RefundStatus {
	switch v {
	case "PROCESSING", "SUCCESS", "REFUNDCLOSE", "CHANGE":
		return RefundStatus{v}
	default:
		return RefundStatus{}
	}
}

// String 实现 Stringer 接口
func (rs RefundStatus) String() string {
	return rs.v
}

// IsValid 当该值有效(非空)时返回 true
func (rs RefundStatus) IsValid() bool {
	return rs.v != ""
}

// ParseSignType parse 签名类型
func ParseSignType(v string) SignType {
	switch v {
	case "MD5", "HMAC-SHA256":
		return SignType{v}
	default:
		return SignType{}
	}
}

// String 实现 Stringer 接口
func (st SignType) String() string {
	return st.v
}

// IsValid 当该值有效(非空)时返回 true
func (st SignType) IsValid() bool {
	return st.v != ""
}

// String 实现 Stringer 接口
func (v Version) String() string {
	return v.v
}
