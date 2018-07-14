package mch

import (
	"context"
	"errors"
	"github.com/huangjunwen/wxdriver/conf"
	"net/http"
	"time"
)

var (
	ErrOrderQueryMissingID       = errors.New("Missing transaction_id/out_trade_no in OrderQueryRequest")
	ErrOrderQueryNoOutTradeNo    = errors.New("No out_trade_no is returned from OrderQueryResponse")
	ErrOrderQueryNoTradeState    = errors.New("No trade_state is returned from OrderQueryResponse")
	ErrOrderQueryNoTransactionID = errors.New("No transaction_id is returned from OrderQueryResponse")
	ErrOrderQueryNoOpenID        = errors.New("No openid is returned from OrderQueryResponse")
	ErrOrderQueryNoTradeType     = errors.New("No trade_type is returned from OrderQueryResponse")
	ErrOrderQueryNoBankType      = errors.New("No bank_type is returned from OrderQueryResponse")
	ErrOrderQueryNoTimeEnd       = errors.New("No time_end is returned from OrderQueryResponse")
	ErrOrderQueryNoTotalFee      = errors.New("No total_fee is returned from OrderQueryResponse")
	ErrOrderQueryNoCashFee       = errors.New("No cash_fee is returned from OrderQueryResponse")
)

// OrderQueryRequest 为查询订单接口请求
type OrderQueryRequest struct {
	// ----- 必填字段 -----
	// 以下二选一
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

	// ----- 支付完成后(SUCCESS, REFUND) 必返回字段 -----
	TransactionID string    // transaction_id String(32) 微信支付订单号
	OpenID        string    // openid String(128) 用户标识
	TradeType     TradeType // trade_type String(16) 交易类型
	BankType      string    // bank_type String(16) 付款银行
	TimeEnd       time.Time // time_end String(14) 订单支付时间
	TotalFee      uint64    // total_fee Int 标价金额
	CashFee       uint64    // cash_fee Int 现金支付金额

	// ----- 支付完成后可能返回的字段 -----
	FeeType     string // fee_type String(16) 标价币种
	CashFeeType string // cash_fee_type String(16) 现金支付币种
	Rate        uint64 // rate String(16) 汇率 标价币种与支付币种兑换比例乘以10^8
	// TODO: 优惠金额字段
	// TODO: 这里涉及不同 Version 时返回的字段不一样，应当用统一的结构存储

	// ----- 其它字段 -----
	DeviceInfo     string // device_info String(32) 设备号
	TradeStateDesc string // trade_state_desc String(256) 交易状态描述
	IsSubscribe    string // is_subscribe String(1) Y/N 是否关注公众账号 仅在公众账号类型支付有效
	Attach         string // attach String(127) 附加数据
}

func orderQuery(ctx context.Context, config conf.MchConfig, req *OrderQueryRequest, options *Options) (*OrderQueryResponse, error) {
	// req -> reqXML
	reqXML := MchXML{}
	if req.TransactionID != "" {
		reqXML.fillString(req.TransactionID, "transaction_id")
	} else if req.OutTradeNo != "" {
		reqXML.fillString(req.OutTradeNo, "out_trade_no")
	} else {
		return nil, ErrOrderQueryMissingID
	}

	// reqXML -> respXML
	respXML, err := PostMchXML(ctx, config, "/pay/orderquery", reqXML, options)
	if err != nil {
		return nil, err
	}

	// respXML -> resp
	resp := OrderQueryResponse{
		MchXML: respXML,
	}
	respXML.extractString(&resp.OutTradeNo, "out_trade_no", &err)
	respXML.extractTradeState(&resp.TradeState, "trade_state", &err)
	respXML.extractString(&resp.TransactionID, "transaction_id", &err)
	respXML.extractString(&resp.OpenID, "openid", &err)
	respXML.extractTradeType(&resp.TradeType, "trade_type", &err)
	respXML.extractString(&resp.BankType, "bank_type", &err)
	respXML.extractTimeCompact(&resp.TimeEnd, "time_end", &err)
	respXML.extractUint64(&resp.TotalFee, "total_fee", &err)
	respXML.extractString(&resp.FeeType, "fee_type", &err)
	respXML.extractUint64(&resp.CashFee, "cash_fee", &err)
	respXML.extractString(&resp.CashFeeType, "cash_fee_type", &err)
	respXML.extractUint64(&resp.Rate, "rate", &err)
	respXML.extractString(&resp.DeviceInfo, "device_info", &err)
	respXML.extractString(&resp.TradeStateDesc, "trade_state_desc", &err)
	respXML.extractString(&resp.IsSubscribe, "is_subscribe", &err)
	respXML.extractString(&resp.Attach, "attach", &err)
	if err != nil {
		return nil, err
	}

	// 检查返回字段
	if resp.OutTradeNo == "" {
		return nil, ErrOrderQueryNoOutTradeNo
	}
	if !resp.TradeState.IsValid() {
		return nil, ErrOrderQueryNoTradeState
	}

	switch resp.TradeState {
	case TradeStateSUCCESS, TradeStateREFUND:
		if resp.TransactionID == "" {
			return nil, ErrOrderQueryNoTransactionID
		}
		if resp.OpenID == "" {
			return nil, ErrOrderQueryNoOpenID
		}
		if !resp.TradeType.IsValid() {
			return nil, ErrOrderQueryNoTradeType
		}
		if resp.BankType == "" {
			return nil, ErrOrderQueryNoBankType
		}
		if resp.TimeEnd.IsZero() {
			return nil, ErrOrderQueryNoTimeEnd
		}
		if resp.TotalFee == 0 {
			return nil, ErrOrderQueryNoTotalFee
		}
		// CashFee 说不定可能为 0 （完全用优惠金支付），如果为 0，则需要额外检查 xml 是否
		// 有该项返回，若有返回则则不报错
		if resp.CashFee == 0 {
			if respXML["cash_fee"] == "" {
				return nil, ErrOrderQueryNoCashFee
			}
		}
	}

	return &resp, nil
}

// OrderQuery 查询订单接口
func OrderQuery(ctx context.Context, config conf.MchConfig, req *OrderQueryRequest, opts ...Option) (*OrderQueryResponse, error) {
	options, err := NewOptions(opts...)
	if err != nil {
		return nil, err
	}
	return orderQuery(ctx, config, req, options)
}

// OrderNotify 创建一个处理支付结果通知的 http.Handler; 传入 handler 的参数包括上下文和查询订单接口返回的 Response；handler
// 处理过后若成功应该返回 nil，若失败则应该返回一个非 nil error 对象，该 error 的 String() 将会返回给外部
//
// NOTE：请使用与在统一下单一样的签名类型，否则签名会可能不通过
func OrderNotify(handler func(context.Context, *OrderQueryResponse) error, selector conf.MchConfigSelector, options *Options) http.Handler {

	return HandleSignedMchXML(func(ctx context.Context, x MchXML) error {
		// 这里再次发起查询有以下原因
		// 1. 回调所带的参数虽然与查询接口返回的几乎一致，但依据文档显示回调里好像没有包含 trade_state，
		//    再次发起查询能与主动查询保持一致
		// 2. 回调虽然带有签名，但万一 key 泄漏则任何人都可以伪造；主动发起查询则能多一层防护
		resp, err := orderQuery(ctx, selector.SelectMch(x["appid"], x["mch_id"]), &OrderQueryRequest{
			TransactionID: x["transaction_id"],
			OutTradeNo:    x["out_trade_no"],
		}, options)
		if err != nil {
			return err
		}
		return handler(ctx, resp)
	}, selector, options)

}
