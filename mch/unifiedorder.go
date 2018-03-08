package mch

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUnifiedOrderMissingOutTradeNo     = errors.New("Missing out_trade_no in UnifiedOrderRequest")
	ErrUnifiedOrderMissingTotalFee       = errors.New("Missing total_fee in UnifiedOrderRequest")
	ErrUnifiedOrderMissingBody           = errors.New("Missing body in UnifiedOrderRequest")
	ErrUnifiedOrderMissingSpbillCreateIp = errors.New("Missing spbill_create_ip in UnifiedOrderRequest")
	ErrUnifiedOrderMissingNotifyUrl      = errors.New("Missing notify_url in UnifiedOrderRequest")
	ErrUnifiedOrderMissingTradeType      = errors.New("Missing trade_type in UnifiedOrderRequest")
	ErrUnifiedOrderMissingOpenID         = errors.New("Missing openid in UnifiedOrderRequest since trade_type is JSAPI")
	ErrUnifiedOrderMissingProductID      = errors.New("Missing product_id in UnifiedOrderRequest since trade_type is NATIVE")
	ErrUnifiedOrderBadTradeType          = errors.New("Bad trade_type is returned from UnifiedOrderResponse")
	ErrUnifiedOrderNoPrepayID            = errors.New("No prepay_id is returned from UnifiedOrderResponse")
	ErrUnifiedOrderNoCodeUrl             = errors.New("No code_url is returned from UnifiedOrderResponse")
	ErrUnifiedOrderNoMWebUrl             = errors.New("No mweb_url is returned from UnifiedOrderResponse")
)

// UnifiedOrderRequest 为统一下单接口请求
type UnifiedOrderRequest struct {
	// ----- 必填字段 -----
	OutTradeNo     string    // out_trade_no String(32) 商户系统内部订单号 同一个商户号下唯一
	TotalFee       uint64    // total_fee Int 标价金额 单位为分
	Body           string    // body String(128) 商品描述 <商场名>-<商品名>
	SpbillCreateIp string    // spbill_create_ip String(16) 终端IP
	NotifyUrl      string    // notify_url String(256) 通知地址
	TradeType      TradeType // trade_type String(16) 交易类型

	// ----- 特定条件必填字段 -----
	OpenID string // openid String(128) 用户标识 trade_type 为 JSAPI 时必填

	// ----- 选填字段 -----
	ProductID  string    // product_id String(32) 商户自定义商品 ID trade_type 为 NATIVE 时必传
	DeviceInfo string    // device_info String(32) 设备号
	Detail     string    // detail String(6000) 商品详情
	Attach     string    // attach String(127) 附加数据
	FeeType    string    // fee_type String(16) 标价币种
	TimeStart  time.Time // time_start String(14) 交易起始时间 格式如 20091225091010
	TimeExpire time.Time // time_expire String(14) 交易结束时间
	GoodsTag   string    // goods_tag String(32) 订单优惠标记
	LimitPay   string    // limit_pay String(32)指定支付方式

	// TOOD: 单品优惠
}

// UnifiedOrderResponse 为统一下单接口响应
type UnifiedOrderResponse struct {
	// ----- 原始数据 -----
	MchXML MchXML

	// ----- 必返回字段 -----
	TradeType TradeType // trade_type String(16) 交易类型
	PrepayID  string    // prepay_id String(64) 预支付交易会话标识

	// ----- 特定条件返回字段 -----
	CodeUrl string // code_url String(64) 二维码链接 trade_type 为 NATIVE 时有返回
	MWebUrl string // mweb_url String(64) 支付跳转链接 trade_type 为 MWEB 时有返回 可通过访问该url来拉起微信客户端

	// ----- 其它字段
	DeviceInfo string // device_info String(32) 设备号

}

// UnifiedOrder 统一下单接口
func UnifiedOrder(ctx context.Context, config Config, req *UnifiedOrderRequest, opts ...Option) (*UnifiedOrderResponse, error) {
	// req -> reqXML
	reqXML := MchXML{}
	if req.OutTradeNo == "" {
		return nil, ErrUnifiedOrderMissingOutTradeNo
	} else {
		reqXML.fillString(req.OutTradeNo, "out_trade_no")
	}

	if req.TotalFee == 0 {
		return nil, ErrUnifiedOrderMissingTotalFee
	} else {
		reqXML.fillUint64(req.TotalFee, "total_fee")
	}

	if req.Body == "" {
		return nil, ErrUnifiedOrderMissingBody
	} else {
		reqXML.fillString(req.Body, "body")
	}

	if req.SpbillCreateIp == "" {
		return nil, ErrUnifiedOrderMissingSpbillCreateIp
	} else {
		reqXML.fillString(req.SpbillCreateIp, "spbill_create_ip")
	}

	if req.NotifyUrl == "" {
		return nil, ErrUnifiedOrderMissingNotifyUrl
	} else {
		reqXML.fillString(req.NotifyUrl, "notify_url")
	}

	if !req.TradeType.IsValid() {
		return nil, ErrUnifiedOrderMissingTradeType
	} else {
		reqXML.fillStringer(req.TradeType, "trade_type")
	}

	if req.TradeType == TradeTypeJSAPI && req.OpenID == "" {
		return nil, ErrUnifiedOrderMissingOpenID
	}
	if req.OpenID != "" {
		reqXML.fillString(req.OpenID, "openid")
	}

	if req.ProductID != "" {
		reqXML.fillString(req.ProductID, "product_id")
	}
	if req.DeviceInfo != "" {
		reqXML.fillString(req.DeviceInfo, "device_info")
	}
	if req.Detail != "" {
		reqXML.fillString(req.Detail, "detail")
	}
	if req.Attach != "" {
		reqXML.fillString(req.Attach, "attach")
	}
	if req.FeeType != "" {
		reqXML.fillString(req.FeeType, "fee_type")
	}
	if !req.TimeStart.IsZero() {
		reqXML.fillTimeCompact(req.TimeStart, "time_start")
	}
	if !req.TimeExpire.IsZero() {
		reqXML.fillTimeCompact(req.TimeExpire, "time_expire")
	}
	if req.GoodsTag != "" {
		reqXML.fillString(req.GoodsTag, "goods_tag")
	}
	if req.LimitPay != "" {
		reqXML.fillString(req.LimitPay, "limit_pay")
	}

	// reqXML -> respXML
	respXML, err := postMchXML(ctx, config, "/pay/unifiedorder", reqXML, opts)
	if err != nil {
		return nil, err
	}

	// respXML -> resp
	resp := UnifiedOrderResponse{
		MchXML: respXML,
	}
	respXML.extractTradeType(&resp.TradeType, "trade_type", &err)
	respXML.extractString(&resp.PrepayID, "prepay_id", &err)
	respXML.extractString(&resp.CodeUrl, "code_url", &err)
	respXML.extractString(&resp.MWebUrl, "mweb_url", &err)
	respXML.extractString(&resp.DeviceInfo, "device_info", &err)
	if err != nil {
		return nil, err
	}

	// 检查 resp
	if !resp.TradeType.IsValid() || req.TradeType != resp.TradeType {
		return nil, ErrUnifiedOrderBadTradeType
	}
	if resp.PrepayID == "" {
		return nil, ErrUnifiedOrderNoPrepayID
	}
	if resp.TradeType == TradeTypeNATIVE && resp.CodeUrl == "" {
		return nil, ErrUnifiedOrderNoCodeUrl
	}
	if resp.TradeType == TradeTypeMWEB && resp.MWebUrl == "" {
		return nil, ErrUnifiedOrderNoMWebUrl
	}

	return &resp, nil

}
