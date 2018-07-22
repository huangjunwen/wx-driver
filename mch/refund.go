package mch

import (
	"context"
	"errors"

	"github.com/huangjunwen/wxdriver/conf"
)

var (
	ErrRefundMissingOutRefundNo = errors.New("Missing out_refund_no in RefundRequest")
	ErrRefundMissingID          = errors.New("Missing transaction_id/out_trade_no in RefundRequest")
	ErrRefundMissingTotalFee    = errors.New("Missing total_fee in RefundRequest")
	ErrRefundMissingRefundFee   = errors.New("Missing refund_fee in RefundRequest")
	ErrRefundNoTransactionID    = errors.New("No transaction_id is returned from RefundResponse")
	ErrRefundNoOutTradeNo       = errors.New("No out_trade_no is returned from RefundResponse")
	ErrRefundNoRefundID         = errors.New("No refund_id is returned from RefundResponse")
	ErrRefundNoOutRefundNo      = errors.New("No out_refund_no is returned from RefundResponse")
	ErrRefundNoTotalFee         = errors.New("No total_fee is returned from RefundResponse")
	ErrRefundNoCashFee          = errors.New("No cash_fee is returned from RefundResponse")
	ErrRefundNoRefundFee        = errors.New("No refund_fee is returned from RefundResponse")
	ErrRefundNoCashRefundFee    = errors.New("No cash_refund_fee is returned from RefundResponse")
)

// RefundRequest 为申请退款接口请求
type RefundRequest struct {
	// ----- 必填字段 -----
	OutRefundNo string // out_refund_no String(64) 商户退款单号，商户系统内部唯一

	// 以下二选一
	TransactionID string // transaction_id String(32) 微信订单号
	OutTradeNo    string // out_trade_no String(32) 商户订单号，商户系统内部唯一

	// 金额字段
	TotalFee  uint64 // total_fee Int 标价金额 单位为分
	RefundFee uint64 // refund_fee Int 退款金额 (NOTE: 此退款金额的币种跟标价金额是一致的)

	// ----- 选填字段 -----
	RefundFeeType string // refund_fee_type String(8) 退款币种，必须与标价金额一致
	RefundDesc    string // refund_desc String(80) 退款原因

}

// RefundRequest 为申请退款接口响应
type RefundResponse struct {
	// ----- 原始数据 -----
	MchXML MchXML

	// ----- 必返回字段 -----
	TransactionID string // transaction_id String(32) 微信订单号
	OutTradeNo    string // out_trade_no String(32) 商户订单号
	RefundID      string // refund_id String(32) 微信退款单号
	OutRefundNo   string // out_refund_no String(64) 商户退款单号
	TotalFee      uint64 // total_fee Int 标价金额
	CashFee       uint64 // cash_fee Int 现金支付金额
	RefundFee     uint64 // refund_fee Int 退款金额
	CashRefundFee uint64 // cash_refund_fee Int 现金退款金额

	// ----- 选传字段 -----
	FeeType           string // fee_type String(16) 标价币种
	CashFeeType       string // cash_fee_type String(16) 现金支付币种
	Rate              uint64 // rate String(16) 汇率 标价币种与支付币种兑换比例乘以10^8
	RefundFeeType     string // refund_fee_type String(8) 退款币种
	CashRefundFeeType string // cash_refund_fee_type String(8) 现金退款金额币种

	// TODO: 优惠金额字段
}

// Refund 申请退款接口，该接口需要客户端证书的 client
func Refund(ctx context.Context, config conf.MchConfig, req *RefundRequest, opts ...Option) (*RefundResponse, error) {
	options, err := NewOptions(opts...)
	if err != nil {
		return nil, err
	}

	// req -> reqXML
	reqXML := MchXML{}

	if req.OutRefundNo != "" {
		reqXML.fillString(req.OutRefundNo, "out_refund_no")
	} else {
		return nil, ErrRefundMissingOutRefundNo
	}

	if req.TransactionID != "" {
		reqXML.fillString(req.TransactionID, "transaction_id")
	} else if req.OutTradeNo != "" {
		reqXML.fillString(req.OutTradeNo, "out_trade_no")
	} else {
		return nil, ErrRefundMissingID
	}

	if req.TotalFee != 0 {
		reqXML.fillUint64(req.TotalFee, "total_fee")
	} else {
		return nil, ErrRefundMissingTotalFee
	}

	if req.RefundFee != 0 {
		reqXML.fillUint64(req.RefundFee, "refund_fee")
	} else {
		return nil, ErrRefundMissingRefundFee
	}

	// reqXML -> respXML
	respXML, err := PostMchXML(ctx, config, "/secapi/pay/refund", reqXML, options)
	if err != nil {
		return nil, err
	}

	// respXML -> resp
	resp := RefundResponse{
		MchXML: respXML,
	}
	respXML.extractString(&resp.TransactionID, "transaction_id", &err)
	respXML.extractString(&resp.OutTradeNo, "out_trade_no", &err)
	respXML.extractString(&resp.RefundID, "refund_id", &err)
	respXML.extractString(&resp.OutRefundNo, "out_refund_no", &err)
	respXML.extractUint64(&resp.TotalFee, "total_fee", &err)
	respXML.extractUint64(&resp.CashFee, "cash_fee", &err)
	respXML.extractUint64(&resp.RefundFee, "refund_fee", &err)
	respXML.extractUint64(&resp.CashRefundFee, "cash_refund_fee", &err)
	respXML.extractString(&resp.FeeType, "fee_type", &err)
	respXML.extractString(&resp.CashFeeType, "cash_fee_type", &err)
	respXML.extractUint64(&resp.Rate, "rate", &err)
	respXML.extractString(&resp.RefundFeeType, "refund_fee_type", &err)
	respXML.extractString(&resp.CashRefundFeeType, "cash_refund_fee_type", &err)
	if err != nil {
		return nil, err
	}

	// 检查返回字段
	if resp.TransactionID == "" {
		return nil, ErrRefundNoTransactionID
	}
	if resp.OutTradeNo == "" {
		return nil, ErrRefundNoOutTradeNo
	}
	if resp.RefundID == "" {
		return nil, ErrRefundNoRefundID
	}
	if resp.OutRefundNo == "" {
		return nil, ErrRefundNoOutRefundNo
	}
	if resp.TotalFee == 0 {
		return nil, ErrRefundNoTotalFee
	}
	if resp.RefundFee == 0 {
		return nil, ErrRefundNoRefundFee
	}
	// CashFee/CashRefundFee 说不定可能为 0 （完全用优惠金支付），如果为 0，则需要额外检查 xml 是否
	// 有该项返回，若有返回则则不报错
	if resp.CashFee == 0 {
		if respXML["cash_fee"] == "" {
			return nil, ErrRefundNoCashFee
		}
	}
	if resp.CashRefundFee == 0 {
		if respXML["cash_refund_fee"] == "" {
			return nil, ErrRefundNoCashRefundFee
		}
	}

	return &resp, nil

}
