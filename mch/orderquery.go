package mch

import (
	"context"
	"errors"
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

func newOrderQueryResponse(respXML MchXML) (*OrderQueryResponse, error) {
	var err error

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
		resp.TotalFee, err = strconv.ParseUint(respXML["total_fee"], 10, 32)
		if err != nil {
			return nil, err
		}
	}
	if respXML["cash_fee"] != "" {
		resp.CashFee, err = strconv.ParseUint(respXML["cash_fee"], 10, 32)
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
	resp, err := newOrderQueryResponse(respXML)
	if err != nil {
		return nil, err
	}

	return resp, nil

}
