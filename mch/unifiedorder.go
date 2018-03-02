package mch

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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
)

// UnifiedOrderRequest 为统一下单接口请求
type UnifiedOrderRequest struct {
	// ----- 必填字段 -----
	OutTradeNo     string    // out_trade_no String(32) 商户系统内部订单号 同一个商户号下唯一
	TotalFee       uint      // total_fee Int 标价金额 单位为分
	Body           string    // body String(128) 商品描述 <商场名>-<商品名>
	SpbillCreateIp string    // spbill_create_ip String(16) 终端IP
	NotifyUrl      string    // notify_url String(256) 通知地址
	TradeType      TradeType // trade_type String(16) 交易类型

	// ----- 特定条件必填字段 -----
	OpenID    string // openid String(128) 用户标识 trade_type 为 JSAPI 时必填
	ProductID string // product_id String(32) 商户自定义商品 ID trade_type 为 NATIVE 时必传

	// ----- 选填字段 -----
	DeviceInfo string    // device_info String(32) 设备号
	Detail     string    // detail String(6000) 商品详情
	Attach     string    // attach String(127) 附加数据
	FeeType    string    // fee_type String(16) 标价币种
	TimeStart  time.Time // time_start String(14) 交易起始时间 格式如 20091225091010
	TimeExpire time.Time // time_expire String(14) 交易结束时间
	GoodsTag   string    // goods_tag String(32) 订单优惠标记
	LimitPay   string    // limit_pay String(32)指定支付方式
	// TOOD: SceneInfo
}

// UnifiedOrderResponse 为统一下单接口响应
type UnifiedOrderResponse struct {
	MchResponse
	TradeType TradeType // trade_type String(16) 交易类型
	PrepayID  string    // prepay_id String(64) 预支付交易会话标识
	CodeUrl   string    // code_url String(64) 二维码链接 trade_type 为 NATIVE 时有返回
	MWebUrl   string    // mweb_url String(64) 支付跳转链接 trade_type 为 MWEB 时有返回 可通过访问该url来拉起微信客户端
}

func UnifiedOrder(ctx context.Context, config Configuration, req *UnifiedOrderRequest, options ...Option) (*UnifiedOrderResponse, error) {
	opts, err := NewOptions(options...)
	if err != nil {
		return nil, err
	}

	reqXML := mchXML{}
	if req.OutTradeNo == "" {
		return nil, ErrUnifiedOrderMissingOutTradeNo
	} else {
		reqXML.AddField("out_trade_no", req.OutTradeNo)
	}

	if req.TotalFee == 0 {
		return nil, ErrUnifiedOrderMissingTotalFee
	} else {
		reqXML.AddField("total_fee", strconv.FormatUint(uint64(req.TotalFee), 10))
	}

	if req.Body == "" {
		return nil, ErrUnifiedOrderMissingBody
	} else {
		reqXML.AddField("body", req.Body)
	}

	if req.SpbillCreateIp == "" {
		return nil, ErrUnifiedOrderMissingSpbillCreateIp
	} else {
		reqXML.AddField("spbill_create_ip", req.SpbillCreateIp)
	}

	if req.NotifyUrl == "" {
		return nil, ErrUnifiedOrderMissingNotifyUrl
	} else {
		reqXML.AddField("notify_url", req.NotifyUrl)
	}

	if !req.TradeType.IsValid() {
		return nil, ErrUnifiedOrderMissingTradeType
	} else {
		reqXML.AddField("trade_type", req.TradeType.String())
	}

	if req.TradeType == TradeTypeJSAPI && req.OpenID == "" {
		return nil, ErrUnifiedOrderMissingOpenID
	} else {
		reqXML.AddField("openid", req.OpenID)
	}

	if req.TradeType == TradeTypeNATIVE && req.ProductID == "" {
		return nil, ErrUnifiedOrderMissingProductID
	} else {
		reqXML.AddField("product_id", req.ProductID)
	}

	if req.DeviceInfo != "" {
		reqXML.AddField("device_info", req.DeviceInfo)
	}
	if req.Detail != "" {
		reqXML.AddField("detail", req.Detail)
	}
	if req.Attach != "" {
		reqXML.AddField("attach", req.Attach)
	}
	if req.FeeType != "" {
		reqXML.AddField("fee_type", req.FeeType)
	}
	if !req.TimeStart.IsZero() {
		reqXML.AddField("time_start", req.TimeStart.Format(datetimeLayout))
	}
	if !req.TimeExpire.IsZero() {
		reqXML.AddField("time_expire", req.TimeExpire.Format(datetimeLayout))
	}
	if req.GoodsTag != "" {
		reqXML.AddField("goods_tag", req.GoodsTag)
	}
	if req.LimitPay != "" {
		reqXML.AddField("limit_pay", req.LimitPay)
	}

	respXML := mchXML{}
	err = postMchXML(ctx, config, "https://api.mch.weixin.qq.com/pay/unifiedorder", &reqXML, &respXML, opts)
	if err != nil {
		return nil, err
	}

	resp := UnifiedOrderResponse{}
	err = respXML.EachField(resp.MchResponse.mchXMLIter, func(_ int, fieldName, fieldValue string) error {
		switch fieldName {
		case "trade_type":
			resp.TradeType = ParseTradeType(fieldValue)
			if !resp.TradeType.IsValid() {
				return fmt.Errorf("Unknwon trade type %+q", fieldValue)
			}
		case "prepay_id":
			resp.PrepayID = fieldValue
		case "code_url":
			resp.CodeUrl = fieldValue
		case "mweb_url":
			resp.MWebUrl = fieldValue
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return &resp, resp.Error()
	}

	return &resp, nil

}
