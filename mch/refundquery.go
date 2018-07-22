package mch

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/huangjunwen/wxdriver/conf"
)

var (
	ErrRefundQueryMissingID       = errors.New("Missing refund_id/out_refund_no/transaction_id/out_trade_no in RefundQueryRequest")
	ErrRefundQueryNoOutTradeNo    = errors.New("No out_trade_no is returned from RefundQueryResponse")
	ErrRefundQueryNoTransactionID = errors.New("No transaction_id is returned from RefundQueryResponse")
	ErrRefundQueryNoTotalFee      = errors.New("No total_fee is returned from RefundQueryResponse")
	ErrRefundQueryNoCashFee       = errors.New("No cash_fee is returned from RefundQueryResponse")
	ErrRefundQueryNoRefundCount   = errors.New("No refund_count is returned from RefundQueryResponse")
)

// RefundQueryRequest 为查询退款接口请求
type RefundQueryRequest struct {
	// ----- 必填字段 -----
	// 以下四选一，优先级为 refund_id > out_refund_no > transaction_id > out_trade_no
	RefundID      string // refund_id String(32) 微信退款单号
	OutRefundNo   string // out_refund_no String(64) 商户退款单号
	TransactionID string // transaction_id String(32) 微信订单号
	OutTradeNo    string // out_trade_no String(32) 商户订单号

	// ----- 选填字段 -----
	Offset uint // offset Int 偏移量
}

// RefundQueryRequest 为查询退款接口响应
type RefundQueryResponse struct {
	// ----- 原始数据 -----
	MchXML MchXML

	// ----- 必返回字段 -----
	OutTradeNo    string       // out_trade_no String(32) 商户订单号
	TransactionID string       // transaction_id String(32) 微信订单号
	TotalFee      uint64       // total_fee Int 标价金额
	CashFee       uint64       // cash_fee Int 现金支付金额
	RefundCount   uint64       // refund_count 当前返回退款单数
	Refunds       []RefundInfo // 退款单信息

	// ----- 其它字段 -----
	TotalRefundCount uint64 // total_refund_count Int 订单总退款次数
	FeeType          string // fee_type String(16) 标价币种
	CashFeeType      string // cash_fee_type String(16) 现金支付币种
	Rate             uint64 // rate String(16) 汇率 标价币种与支付币种兑换比例乘以10^8
}

// RefundInfo 为查询退款接口响应中的单笔退款单信息
type RefundInfo struct {
	// ----- 必返回字段 -----
	RefundID     string       // refund_id_$n String(32) 微信退款单号
	OutRefundNo  string       // out_refund_no_$n String(64) 商户退款单号
	RefundFee    uint64       // refund_fee_$n Int 退款金额
	RefundStatus RefundStatus // refund_status_$n String(16) 退款状态

	// ----- 其它字段 -----
	RefundChannel     string    // refund_channel_$n String(16) 退款渠道
	RefundAccount     string    // refund_account_$n String(30) 退款资金来源
	RefundRecvAccout  string    // refund_recv_accout_$n String(64) 退款入账账户
	RefundSuccessTime time.Time // refund_success_time_$n String(20) 退款成功时间 (2016-07-25 15:26:26)
}

func refundQuery(ctx context.Context, config conf.MchConfig, req *RefundQueryRequest, options *Options) (*RefundQueryResponse, error) {
	// req -> reqXML
	reqXML := MchXML{}
	if req.RefundID != "" {
		reqXML.fillString(req.RefundID, "refund_id")
	} else if req.OutRefundNo != "" {
		reqXML.fillString(req.OutRefundNo, "out_refund_no")
	} else if req.TransactionID != "" {
		reqXML.fillString(req.TransactionID, "transaction_id")
	} else if req.OutTradeNo != "" {
		reqXML.fillString(req.OutTradeNo, "out_trade_no")
	} else {
		return nil, ErrRefundQueryMissingID
	}

	if req.Offset != 0 {
		reqXML.fillUint64(uint64(req.Offset), "offset")
	}

	// reqXML -> respXML
	respXML, err := PostMchXML(ctx, config, "/pay/refundquery", reqXML, options)
	if err != nil {
		return nil, err
	}

	// respXML -> resp
	resp := RefundQueryResponse{
		MchXML: respXML,
	}
	respXML.extractString(&resp.OutTradeNo, "out_trade_no", &err)
	respXML.extractString(&resp.TransactionID, "transaction_id", &err)
	respXML.extractUint64(&resp.TotalFee, "total_fee", &err)
	respXML.extractUint64(&resp.CashFee, "cash_fee", &err)
	respXML.extractUint64(&resp.RefundCount, "refund_count", &err)
	respXML.extractUint64(&resp.TotalRefundCount, "total_refund_count", &err)
	respXML.extractString(&resp.FeeType, "fee_type", &err)
	respXML.extractString(&resp.CashFeeType, "cash_fee_type", &err)
	respXML.extractUint64(&resp.Rate, "rate", &err)
	if err != nil {
		return nil, err
	}

	if resp.OutTradeNo == "" {
		return nil, ErrRefundQueryNoOutTradeNo
	}
	if resp.TransactionID == "" {
		return nil, ErrRefundQueryNoTransactionID
	}
	if resp.TotalFee == 0 {
		return nil, ErrRefundQueryNoTotalFee
	}
	if resp.CashFee == 0 {
		return nil, ErrRefundQueryNoCashFee
	}
	if resp.RefundCount == 0 {
		// NOTE: 据实际测试，若支付订单没有退款单不会返回一个 0 的 refund_count 而是返回错误，
		// 所以这里直接检查为 0 就好
		return nil, ErrRefundQueryNoRefundCount
	}

	resp.Refunds = make([]RefundInfo, resp.RefundCount)
	for i := 0; i < int(resp.RefundCount); i++ {
		ri := &resp.Refunds[i]
		idx := i + int(req.Offset)
		respXML.extractString(&ri.RefundID, fmt.Sprintf("refund_id_%d", idx), &err)
		respXML.extractString(&ri.OutRefundNo, fmt.Sprintf("out_refund_no_%d", idx), &err)
		respXML.extractUint64(&ri.RefundFee, fmt.Sprintf("refund_fee_%d", idx), &err)
		respXML.extractRefundStatus(&ri.RefundStatus, fmt.Sprintf("refund_status_%d", idx), &err)
		respXML.extractString(&ri.RefundChannel, fmt.Sprintf("refund_channel_%d", idx), &err)
		respXML.extractString(&ri.RefundAccount, fmt.Sprintf("refund_account_%d", idx), &err)
		respXML.extractString(&ri.RefundRecvAccout, fmt.Sprintf("refund_recv_accout_%d", idx), &err)
		respXML.extractTime(&ri.RefundSuccessTime, fmt.Sprintf("refund_success_time_%d", idx), "2006-01-02 15:04:05", &err)
		if err != nil {
			return nil, err
		}

		if ri.RefundID == "" {
			return nil, fmt.Errorf("No refund_id_%d is returned from RefundQueryResponse", idx)
		}
		if ri.OutRefundNo == "" {
			return nil, fmt.Errorf("No out_refund_no_%d is returned from RefundQueryResponse", idx)
		}
		if ri.RefundFee == 0 {
			// NOTE: refund_fee 总是非 0
			return nil, fmt.Errorf("No refund_fee_%d is returned from RefundQueryResponse", idx)
		}
		if !ri.RefundStatus.IsValid() {
			return nil, fmt.Errorf("No refund_status_%d is returned from RefundQueryResponse", idx)
		}
	}

	return &resp, nil
}

// RefundQuery 查询退款
func RefundQuery(ctx context.Context, config conf.MchConfig, req *RefundQueryRequest, opts ...Option) (*RefundQueryResponse, error) {
	options, err := NewOptions(opts...)
	if err != nil {
		return nil, err
	}
	return refundQuery(ctx, config, req, options)
}

// RefundNotify 创建一个处理退款结果通知的 http.Handler; 传入 handler 的参数包括上下文和查询退款接口返回的 Response；handler
// 处理过后若成功应该返回 nil，若失败则应该返回一个非 nil error 对象，该 error 的 String() 将会返回给外部
func RefundNotify(handler func(context.Context, *RefundQueryResponse) error, selector conf.MchConfigSelector, options *Options) http.Handler {

	return HandleMchXML(func(ctx context.Context, x MchXML) error {
		config, err := selector.SelectMch(x["appid"], x["mch_id"])
		if err != nil {
			return err
		}
		if config == nil {
			return errors.New("Unknown app or mch")
		}

		x1, err := DecryptMchXML(config.WechatMchKey(), x["req_info"])
		if err != nil {
			return errors.New("Bad xml data")
		}

		// 这里再次发起查询，原因与 OrderNotify 一样
		resp, err := refundQuery(ctx, config, &RefundQueryRequest{
			RefundID: x1["refund_id"],
		}, options)
		if err != nil {
			return err
		}
		return handler(ctx, resp)

	}, options)

}
