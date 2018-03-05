package mch

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"
)

var (
	ErrOrderQueryMissingID        = errors.New("Missing transaction_id or out_trade_no in OrderQueryRequest")
	ErrOrderQueryUnknownTradType  = errors.New("Unknwon trade_type in OrderQueryResponse")
	ErrOrderQueryUnknownTradState = errors.New("Unknwon trade_state in OrderQueryResponse")
)

// OrderQueryRequest 为查询订单接口请求
type OrderQueryRequest struct {
	// ----- 以下二选一 -----
	TransactionID string // transaction_id String(32) 微信支付订单号 建议优先使用
	OutTradeNo    string // out_trade_no String(32) 商户系统内部订单号 同一个商户号下唯一
}

// OrderQueryResponse 为查询订单接口响应
type OrderQueryResponse struct {
	// ----- 原始数据 -----
	MchXML MchXML

	// ----- 必返回字段 -----
	OutTradeNo string     // out_trade_no String(32) 商户系统内部订单号 同一个商户号下唯一
	TradeState TradeState // trade_state String(32) 交易状态

	// ----- 支付完成后返回字段 -----
	TransactionID string    // transaction_id String(32) 微信支付订单号
	OpenID        string    // openid String(128) 用户标识
	TradeType     TradeType // trade_type String(16) 交易类型
	BankType      string    // bank_type String(16) 付款银行
	TimeEnd       time.Time // time_end String(14) 订单支付时间
	// 金额字段
	TotalFee    uint64 // total_fee Int 标价金额 单位为分
	FeeType     string // fee_type String(16) 标价币种，默认 CNY
	CashFee     uint64 // cash_fee Int 现金支付金额 cash_fee <= total_fee 当有立减/折扣/代金券等时部分金额抵扣
	CashFeeType string // cash_fee_type String(16) 现金支付币种，默认 CNY，境外支付时该值可能跟 fee_type 不一致
	Rate        uint64 // rate String(16) 汇率 标价币种与支付币种兑换比例乘以10^8
	// TODO: 优惠金额字段
	// TODO: 这里涉及不同 Version 时返回的字段不一样，应当用统一的结构存储

	// ----- 其它字段 -----
	DeviceInfo     string // device_info String(32) 设备号
	TradeStateDesc string // trade_state_desc String(256) 交易状态描述
	IsSubscribe    string // is_subscribe String(1) Y/N 是否关注公众账号 仅在公众账号类型支付有效
	Attach         string // attach String(127) 附加数据
}

// OrderQuery 查询订单接口
func OrderQuery(ctx context.Context, config Configuration, req *OrderQueryRequest, opts ...Option) (*OrderQueryResponse, error) {
	options, err := NewOptions(opts...)
	if err != nil {
		return nil, err
	}

	// req -> reqXML
	reqXML := MchXML{}
	if req.TransactionID != "" {
		reqXML["transaction_id"] = req.TransactionID
	} else if req.OutTradeNo != "" {
		reqXML["out_trade_no"] = req.OutTradeNo
	} else {
		return nil, ErrOrderQueryMissingID
	}

	// reqXML -> respXML
	respXML, err := postMchXML(ctx, config, "/pay/orderquery", reqXML, options)
	if err != nil {
		return nil, err
	}

	// respXML -> resp
	resp := &OrderQueryResponse{
		MchXML: respXML,
	}

	resp.OutTradeNo = respXML["out_trade_no"]
	resp.TradeState = ParseTradeState(respXML["trade_state"])
	if !resp.TradeState.IsValid() {
		return nil, ErrOrderQueryUnknownTradState
	}

	// 以下全都是可选字段

	resp.TransactionID = respXML["transaction_id"]
	resp.OpenID = respXML["openid"]
	resp.TradeType = ParseTradeType(respXML["trade_type"])
	resp.BankType = respXML["bank_type"]
	if respXML["time_end"] != "" {
		resp.TimeEnd, err = time.Parse(DatetimeLayout, respXML["time_end"])
		if err != nil {
			return nil, err
		}
	}

	if respXML["total_fee"] != "" {
		resp.TotalFee, err = strconv.ParseUint(respXML["total_fee"], 10, 64)
		if err != nil {
			return nil, err
		}
	}
	if respXML["cash_fee"] != "" {
		resp.CashFee, err = strconv.ParseUint(respXML["cash_fee"], 10, 64)
		if err != nil {
			return nil, err
		}
	}
	resp.FeeType = respXML["fee_type"]
	resp.CashFeeType = respXML["cash_fee_type"]
	if respXML["rate"] != "" {
		resp.Rate, err = strconv.ParseUint(respXML["rate"], 10, 64)
		if err != nil {
			return nil, err
		}
	}

	resp.DeviceInfo = respXML["device_info"]
	resp.TradeStateDesc = respXML["trade_state_desc"]
	resp.IsSubscribe = respXML["is_subscribe"]
	resp.Attach = respXML["attach"]

	return resp, nil

}

// OrderNotify 创建一个处理支付结果通知的 http.Handler，NOTE：请使用与在统一下单一样的 options, 否则例如统一下单
// 使用了 HMAC-SHA256 签名，而这里没有，则验证签名有可能会不通过
func OrderNotify(handler func(context.Context, *OrderQueryResponse) error, selector ConfigurationSelector, opts ...Option) http.Handler {

	return handleSignedMchXML(func(ctx context.Context, x MchXML) error {
		// 这里再次发起查询有以下原因
		// 1. 回调所带的参数虽然与查询接口返回的几乎一致，但依据文档显示回调里好像没有包含 trade_state，
		//    再次发起查询能与主动查询保持一致
		// 2. 回调虽然带有签名，但万一 key 泄漏则任何人都可以伪造；主动发起查询则能多一层防护
		resp, err := OrderQuery(ctx, selector.Select(x["appid"], x["mch_id"]), &OrderQueryRequest{
			TransactionID: x["transaction_id"],
			OutTradeNo:    x["out_trade_no"],
		}, opts...)
		if err != nil {
			return err
		}

		// 依据返回的结果执行 handler
		return handler(ctx, resp)

	}, selector, MustOptions(opts...))

}
