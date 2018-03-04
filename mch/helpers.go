package mch

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"github.com/huangjunwen/wxdriver"
	"net/http"
)

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
		signType = SignTypeMD5
	}

	// 添加公共字段
	reqXML["appid"] = config.WechatAppID()
	reqXML["mch_id"] = config.WechatPayMchID()
	reqXML["sign_type"] = signType.String()
	reqXML["nonce_str"] = wxdriver.NonceStr(16) // 32 位以内
	reqXML["sign"] = reqXML.Sign(signType, config.WechatPayKey())

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
	sign := respXML.Sign(signType, config.WechatPayKey())
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

// handleMchXML 处理 mch xml 回调
func handleMchXML(handler func(context.Context, MchXML) error, selector ConfigurationSelector, verifySign bool) http.Handler {

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

		// 提取 appID 和 mchID 以及查找对应配置
		appID := reqXML["appid"]
		mchID := reqXML["mch_id"]
		config := selector.Select(appID, mchID)
		if config == nil {
			writeResponse(false, "")
			return
		}

		// 若需要验证签名
		if verifySign {
			// 提取签名方式
			signType := SignTypeInvalid
			if reqXML["sign_type"] != "" {
				signType = ParseSignType(reqXML["sign_type"])
				if !signType.IsValid() {
					writeResponse(false, "")
					return
				}
			}
			if signType == SignTypeInvalid {
				signType = SignTypeMD5
			}

			// 验证签名
			sign := reqXML.Sign(signType, config.WechatPayKey())
			suppliedSign := reqXML["sign"]
			if suppliedSign == "" || suppliedSign != sign {
				writeResponse(false, "")
				return
			}
		}

		// 执行
		if err := handler(r.Context(), reqXML); err != nil {
			writeResponse(false, err.Error())
			return
		}
		writeResponse(true, "")

	})
}
