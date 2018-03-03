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

type OrderQueryRequest struct {
	// ----- 以下二选一 -----
	TransactionID string // transaction_id String(32) 微信支付订单号 建议优先使用
	OutTradeNo    string // out_trade_no String(32) 商户系统内部订单号 同一个商户号下唯一
}

type OrderQueryResponse struct {
	MchResponse
	// ----- 必传字段 -----
	OutTradeNo     string     // out_trade_no String(32) 商户系统内部订单号 同一个商户号下唯一
	TradeState     TradeState // trade_state String(32) 交易状态
	TradeStateDesc string     // trade_state_desc String(256) 交易状态描述

	// ----- 支付完成后字段 -----
	TransactionID string    // transaction_id String(32) 微信支付订单号
	OpenID        string    // openid String(128) 用户标识
	TradeType     TradeType // trade_type String(16) 交易类型
	BankType      string    // bank_type String(16) 付款银行
	TimeEnd       time.Time // time_end String(14) 订单支付时间
	// 金额字段
	TotalFee    uint32 // total_fee Int 标价金额 单位为分
	FeeType     string // fee_type String(16) 标价币种，默认 CNY
	CashFee     uint32 // cash_fee Int 现金支付金额 cash_fee <= total_fee 当有立减/折扣/代金券等时部分金额抵扣
	CashFeeType string // cash_fee_type String(16) 现金支付币种，默认 CNY，境外支付时该值可能跟 fee_type 不一致
	Rate        uint64 // rate String(16) 汇率 标价币种与支付币种兑换比例乘以10^8
	// 优惠金额字段
	// TODO: 这里涉及不同 Version 时返回的字段不一样，应当用统一的结构存储

	// ----- 选传字段 -----
	DeviceInfo  string // device_info String(32) 设备号
	IsSubscribe string // is_subscribe String(1) Y/N 是否关注公众账号 仅在公众账号类型支付有效
	Attach      string // attach String(127) 附加数据

}

func (resp *OrderQueryResponse) iterCallback(i int, fieldName, fieldValue string) error {
	switch fieldName {
	case "transaction_id":
		resp.TransactionID = fieldValue
	case "out_trade_no":
		resp.OutTradeNo = fieldValue
	case "openid":
		resp.OpenID = fieldValue
	case "trade_type":
		resp.TradeType = ParseTradeType(fieldValue)
		if !resp.TradeType.IsValid() {
			return ErrOrderQueryUnknownTradType
		}
	case "trade_state":
		resp.TradeState = ParseTradeState(fieldValue)
		if !resp.TradeState.IsValid() {
			return ErrOrderQueryUnknownTradState
		}
	case "trade_state_desc":
		resp.TradeStateDesc = fieldValue
	case "bank_type":
		resp.BankType = fieldValue
	case "time_end":
		timeEnd, err := time.Parse(datetimeLayout, fieldValue)
		if err != nil {
			return err
		}
		resp.TimeEnd = timeEnd
	case "total_fee":
		totalFee, err := strconv.ParseUint(fieldValue, 10, 32)
		if err != nil {
			return err
		}
		resp.TotalFee = uint32(totalFee)
	case "fee_type":
		resp.FeeType = fieldValue
	case "cash_fee":
		cashFee, err := strconv.ParseUint(fieldValue, 10, 32)
		if err != nil {
			return err
		}
		resp.CashFee = uint32(cashFee)
	case "cash_fee_type":
		resp.CashFeeType = fieldValue
	case "rate":
		rate, err := strconv.ParseUint(fieldValue, 10, 64)
		if err != nil {
			return err
		}
		resp.Rate = rate
	case "device_info":
		resp.DeviceInfo = fieldValue
	case "is_subscribe":
		resp.IsSubscribe = fieldValue
	case "attach":
		resp.Attach = fieldValue
	}
	return nil
}

func OrderQuery(ctx context.Context, config Configuration, req *OrderQueryRequest, options ...Option) (*OrderQueryResponse, error) {
	opts, err := NewOptions(options...)
	if err != nil {
		return nil, err
	}

	reqXML := mchXML{}
	if req.TransactionID != "" {
		reqXML.AddField("transaction_id", req.TransactionID)
	} else if req.OutTradeNo != "" {
		reqXML.AddField("out_trade_no", req.OutTradeNo)
	} else {
		return nil, ErrOrderQueryMissingID
	}

	respXML := mchXML{}
	err = postMchXML(ctx, config, "/pay/orderquery", &reqXML, &respXML, opts)
	if err != nil {
		return nil, err
	}

	resp := OrderQueryResponse{}
	err = respXML.IterateFields(resp.MchResponse.iterCallback, resp.iterCallback)
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return &resp, resp.Error()
	}

	return &resp, nil

}
