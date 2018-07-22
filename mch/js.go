package mch

import (
	"fmt"
	"strconv"

	"github.com/huangjunwen/wxdriver/conf"
	"github.com/huangjunwen/wxdriver/utils"
)

// JSReqEx 返回拉起微信支付所需的 JS 参数（公众号支付/小程序支付），signType 必须保持和统一下单一致，
// 若均使用 DefaultOptions，可以直接使用 JSReq
func JSReqEx(config conf.MchConfig, prepayID string, signType SignType) map[string]string {
	switch signType {
	case SignTypeMD5, SignTypeHMACSHA256:
	default:
		panic(fmt.Errorf("Bad SignType %+q", signType))
	}
	m := map[string]string{
		"appId":     config.WechatAppID(),
		"timeStamp": strconv.FormatInt(utils.Now().Unix(), 10),
		"nonceStr":  utils.NonceStr(8),
		"package":   fmt.Sprintf("prepay_id=%s", prepayID),
		"signType":  signType.String(),
	}
	m["paySign"] = SignMchXML(MchXML(m), signType, config.WechatMchKey())
	return m
}

// JSReq 返回拉起微信支付所需的 JS 参数（公众号支付/小程序支付），使用 DefaultOptions 中的签名方式
func JSReq(config conf.MchConfig, prepayID string) map[string]string {
	return JSReqEx(config, prepayID, DefaultOptions.SignType())
}
