package mch

import (
	"context"
	"errors"

	"github.com/huangjunwen/wx-driver/conf"
)

var (
	ErrCloseOrderMissingOutTradeNo = errors.New("Missing out_trade_no in CloseOrderRequest")
)

// CloseOrderRequest 是关闭订单接口请求
type CloseOrderRequest struct {
	// ----- 必填字段 -----
	OutTradeNo string // out_trade_no String(32) 商户系统内部订单号 同一个商户号下唯一
}

// CloseOrder 关闭订单接口
func CloseOrder(ctx context.Context, config conf.MchConfig, req *CloseOrderRequest, opts ...Option) error {
	options, err := NewOptions(opts...)
	if err != nil {
		return err
	}

	// req -> reqXML
	reqXML := MchXML{}
	if req.OutTradeNo == "" {
		return ErrCloseOrderMissingOutTradeNo
	} else {
		reqXML.fillString(req.OutTradeNo, "out_trade_no")
	}

	// reqXML -> respXML
	_, err = PostMchXML(ctx, config, "/pay/closeorder", reqXML, options)
	return err
}
