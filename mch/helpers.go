package mch

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"github.com/huangjunwen/wxdriver"
	"hash"
	"net/http"
	"sort"
)

// signMchXML 对 MchXML 进行签名，签名算法见微信支付《安全规范》，signType 为空时默认使用 MD5，
// x 中 sign 字段和空值字段皆不参与签名
func signMchXML(x MchXML, signType SignType, key string) string {
	// 选择 hash
	var h hash.Hash
	switch signType {
	case SignTypeHMACSHA256:
		h = hmac.New(sha256.New, []byte(key))
	default:
		h = md5.New()
	}

	// 排序字段名
	fieldNames := make([]string, 0, len(x))
	for fieldName, _ := range x {
		fieldNames = append(fieldNames, fieldName)
	}
	sort.Strings(fieldNames)

	// 签名
	for _, fieldName := range fieldNames {
		// sign 不参与签名
		if fieldName == "sign" {
			continue
		}

		fieldValue := x[fieldName]
		// 值为空不参与签名
		if fieldValue == "" {
			continue
		}

		h.Write([]byte(fieldName))
		h.Write([]byte("="))
		h.Write([]byte(fieldValue))
		h.Write([]byte("&"))
	}
	h.Write([]byte("key="))
	h.Write([]byte(key))

	// 需要大写
	return fmt.Sprintf("%X", h.Sum(nil))

}

// postMchXML 调用 mch xml 接口，大致过程如下：
//
//   - 添加公共字段 appid/mch_id/mch_id/nonce_str/sign_type
//   - 签名并添加 sign
//   - 调用 api，等待结果或错误
//   - 检查 return_code/return_msg
//   - 验证签名
//   - 验证 appid/mch_id
//   - 检查 result_code
//
// NOTE: 所有参数均不能为空
func postMchXML(ctx context.Context, config Configuration, path string, reqXML MchXML, options *Options) (MchXML, error) {
	// 选择 HTTPClient：options.HTTPClient > DefaultOptions.HTTPClient > wxdriver.DefaultHTTPClient > http.DefaultClient
	client := options.HTTPClient
	if client == nil {
		client = DefaultOptions.HTTPClient
	}
	if client == nil {
		client = wxdriver.DefaultHTTPClient
	}
	if client == nil {
		client = http.DefaultClient
	}

	// 选择 URLBase
	urlBase := options.URLBase
	if urlBase == "" {
		urlBase = DefaultOptions.URLBase
	}
	if urlBase == "" {
		urlBase = URLBaseDefault
	}

	// 选择 SignType
	signType := options.SignType
	if !signType.IsValid() {
		signType = DefaultOptions.SignType
	}
	if !signType.IsValid() {
		signType = SignTypeMD5
	}

	// 添加公共字段
	reqXML["appid"] = config.WechatAppID()
	reqXML["mch_id"] = config.WechatPayMchID()
	reqXML["sign_type"] = signType.String()
	reqXML["nonce_str"] = wxdriver.NonceStr(16) // 32 位以内
	reqXML["sign"] = signMchXML(reqXML, signType, config.WechatPayKey())

	// 编码
	reqBody, err := xml.Marshal(reqXML)
	if err != nil {
		return nil, err
	}

	// 构造请求
	req, err := http.NewRequest("POST", urlBase+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	// 调用!
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// 解码
	respXML := MchXML{}
	if err := xml.NewDecoder(resp.Body).Decode(&respXML); err != nil {
		return nil, err
	}

	// 检查通讯标识 return_code，若失败是没有签名的
	if respXML["return_code"] != "SUCCESS" {
		return nil, fmt.Errorf("Response return_code=%+q return_msg=%+q", respXML["return_code"], respXML["return_msg"])
	}

	// 验证签名
	sign := signMchXML(respXML, signType, config.WechatPayKey())
	suppliedSign := respXML["sign"]
	if suppliedSign == "" || suppliedSign != sign {
		return nil, fmt.Errorf("Response <sign> expect %+q but got %+q", sign, suppliedSign)
	}

	// 验证 appID 和 mchID
	appID := respXML["appid"]
	mchID := respXML["mch_id"]
	if appID != "" && appID != config.WechatAppID() {
		return nil, fmt.Errorf("Response <appid> expect %+q but got %+q", config.WechatAppID(), appID)
	}
	if mchID != "" && mchID != config.WechatPayMchID() {
		return nil, fmt.Errorf("Response <mch_id> expect %+q but got %+q", config.WechatPayMchID(), mchID)
	}

	// 检查业务标识 result_code
	if respXML["result_code"] != "SUCCESS" {
		return nil, fmt.Errorf("Response result_code=%+q err_code=%+q err_code_des=%+q", respXML["result_code"], respXML["err_code"], respXML["err_code_des"])
	}

	// 全部通过
	return respXML, nil

}

// handleMchXML 处理 mch xml 回调，若 handler 返回非 nil error，则该 http.Handler 返回 FAIL return_code 给微信
func handleMchXML(handler func(context.Context, MchXML) error) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		writeResponse := func(success bool, msg string) {
			respXML := MchXML{}
			if success {
				respXML["return_code"] = "SUCCESS"
			} else {
				respXML["return_code"] = "FAIL"
			}
			if msg != "" {
				respXML["return_msg"] = msg
			}
			xml.NewEncoder(w).Encode(respXML)
		}

		// 解码
		reqXML := MchXML{}
		if err := xml.NewDecoder(r.Body).Decode(&reqXML); err != nil {
			writeResponse(false, "")
			return
		}

		// 检查通讯标识 return code，若失败了还回调 ??!
		if reqXML["return_code"] != "SUCCESS" {
			writeResponse(false, "")
			return
		}

		// 执行 handler
		if err := handler(r.Context(), reqXML); err != nil {
			writeResponse(false, err.Error())
			return
		}
		writeResponse(true, "")

	})
}

// handleSignedMchXML 处理带签名的 mch xml 回调，需要传入一个 ConfigurationSelector 用于选择配置
func handleSignedMchXML(handler func(context.Context, MchXML) error, selector ConfigurationSelector, options *Options) http.Handler {

	return handleMchXML(func(ctx context.Context, reqXML MchXML) error {
		// 从 appid 和 mch_id 选择配置（多配置支持）
		config := selector.Select(reqXML["appid"], reqXML["mch_id"])
		if config == nil {
			return fmt.Errorf("Unknown app or mch")
		}

		// 选择签名类型：请求中的 sign_type > options 中的 sign_type > DefaultOptions 中的 sign_type > 默认
		signType := SignTypeInvalid
		if reqXML["sign_type"] != "" {
			signType = ParseSignType(reqXML["sign_type"])
			if !signType.IsValid() {
				return fmt.Errorf("Unknown sign type")
			}
		}
		if !signType.IsValid() {
			signType = options.SignType
		}
		if !signType.IsValid() {
			signType = DefaultOptions.SignType
		}
		if !signType.IsValid() {
			signType = SignTypeMD5
		}

		// 验证签名
		sign := signMchXML(reqXML, signType, config.WechatPayKey())
		suppliedSign := reqXML["sign"]
		if suppliedSign == "" || suppliedSign != sign {
			return fmt.Errorf("Sign error")
		}

		// 通过了，执行 handler
		return handler(ctx, reqXML)

	})

}
