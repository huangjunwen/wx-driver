package mch

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"hash"
	"net/http"
	"sort"

	"github.com/huangjunwen/wx-driver/conf"
	"github.com/huangjunwen/wx-driver/utils"
)

// MchBusinessError 是微信支付业务错误
type MchBusinessError struct {
	// ResultCode 业务结果 SUCCESS/FAIL
	ResultCode string
	// ErrCode 错误代码, 当 ResultCode 为FAIL时返回错误代码
	ErrCode string
	// ErrCodeDes 错误代码描述
	ErrCodeDes string
}

// Error 满足 error 接口
func (err *MchBusinessError) Error() string {
	return fmt.Sprintf(
		"MchBusinessError(result_code=%s err_code=%s err_code_des=%s)",
		err.ResultCode,
		err.ErrCode,
		err.ErrCodeDes,
	)
}

// SignMchXML 对 MchXML 进行签名，签名算法见微信支付《安全规范》，signType 为空时默认使用 MD5，
// x 中 sign 字段和空值字段皆不参与签名；返回的签名字符串为大写
//
// NOTE: 最终用户一般不需要使用该函数
func SignMchXML(x MchXML, signType SignType, mchKey string) string {
	// 选择 hash
	var h hash.Hash
	switch signType {
	case SignTypeHMACSHA256:
		h = hmac.New(sha256.New, []byte(mchKey))
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
	h.Write([]byte(mchKey))

	// 需要大写
	return fmt.Sprintf("%X", h.Sum(nil))

}

// DecryptMchXML 解密一个加密了的 MchXML，目前主要用在退款结果通知，也许未来还有其它地方会用到
//
// NOTE: 最终用户一般不需要使用该函数
func DecryptMchXML(mchKey string, cipherText string) (MchXML, error) {
	// 对商户key做md5，得到32位小写key
	keyMD5 := md5.Sum([]byte(mchKey))
	cipherKey := make([]byte, hex.EncodedLen(md5.Size))
	hex.Encode(cipherKey, keyMD5[:])
	cipherKey = bytes.ToLower(cipherKey)

	// 由该 32 字节 cipherKey 创建 AES-256 cipher
	cipher, err := aes.NewCipher(cipherKey)
	if err != nil {
		return nil, err
	}
	bs := cipher.BlockSize() // 16

	// 用 base64 解码密文，密文长度应当 >0 且为 blocksize 整数倍
	cipherBytes, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return nil, err
	}
	l := len(cipherBytes)
	if l == 0 {
		return nil, fmt.Errorf("Empty cipher text")
	}
	if l%bs != 0 {
		return nil, fmt.Errorf("Cipher text length should be multiply of blocksize")
	}

	// ECB 解密
	plainBytes := make([]byte, l)
	src := cipherBytes
	dst := plainBytes
	for len(src) > 0 {
		cipher.Decrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}

	// pkcs#7 unpadding
	p := int(plainBytes[l-1])
	if p > bs {
		return nil, fmt.Errorf("Padding byte bigger than block size")
	}
	plainBytes = plainBytes[:l-p]

	// xml 解码
	x := MchXML{}
	if err := xml.Unmarshal(plainBytes, &x); err != nil {
		return nil, err
	}

	return x, nil
}

// PostMchXML 调用 mch xml 接口，大致过程如下：
//
//   - 添加公共字段 appid/mch_id/mch_id/nonce_str/sign_type
//   - 签名并添加 sign
//   - 调用 api，等待结果或错误
//   - 检查 return_code/return_msg
//   - 验证签名
//   - 验证 appid/mch_id
//   - 检查 result_code
//
// NOTE: 最终用户一般不需要使用该函数
func PostMchXML(ctx context.Context, config conf.MchConfig, path string, reqXML MchXML, options *Options) (MchXML, error) {
	client := options.Client()
	urlBase := options.URLBase()
	signType := options.SignType()

	// 添加公共字段
	reqXML["appid"] = config.WechatAppID()
	reqXML["mch_id"] = config.WechatMchID()
	reqXML["sign_type"] = signType.String()
	reqXML["nonce_str"] = utils.NonceStr(16) // 32 位以内

	// 签名
	reqXML["sign"] = SignMchXML(reqXML, signType, config.WechatMchKey())

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
	sign := SignMchXML(respXML, signType, config.WechatMchKey())
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
	if mchID != "" && mchID != config.WechatMchID() {
		return nil, fmt.Errorf("Response <mch_id> expect %+q but got %+q", config.WechatMchID(), mchID)
	}

	// 检查业务标识 result_code
	if respXML["result_code"] != "SUCCESS" {
		return nil, &MchBusinessError{
			ResultCode: respXML["result_code"],
			ErrCode:    respXML["err_code"],
			ErrCodeDes: respXML["err_code_des"],
		}
	}

	// 全部通过
	return respXML, nil

}

// HandleMchXML 处理 mch xml 回调，若 handler 返回非 nil error，则该 http.Handler 返回 FAIL return_code 给微信
//
// NOTE: 最终用户一般不需要使用该函数
func HandleMchXML(handler func(context.Context, MchXML) error, options *Options) http.Handler {

	return options.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
			writeResponse(false, "Invalid xml")
			return
		}

		// 检查通讯标识 return code，若失败了还回调 ??!
		if reqXML["return_code"] != "SUCCESS" {
			writeResponse(false, "Failed return_code")
			return
		}

		// 执行 handler
		if err := handler(r.Context(), reqXML); err != nil {
			writeResponse(false, err.Error())
			return
		}
		writeResponse(true, "")

	}))
}

// HandleSignedMchXML 处理带签名的 mch xml 回调，需要传入一个 ConfigurationSelector 用于选择配置
//
// NOTE: 最终用户一般不需要使用该函数
func HandleSignedMchXML(handler func(context.Context, MchXML) error, selector conf.MchConfigSelector, options *Options) http.Handler {

	return HandleMchXML(func(ctx context.Context, reqXML MchXML) error {
		// 从 appid 和 mch_id 选择配置（多配置支持）
		config, err := selector.SelectMch(reqXML["appid"], reqXML["mch_id"])
		if err != nil {
			return err
		}
		if config == nil {
			return errors.New("Unknown app or mch")
		}

		// 选择签名类型：请求中的 sign_type > options 中的 SignType
		signType := SignTypeInvalid
		if reqXML["sign_type"] != "" {
			signType = ParseSignType(reqXML["sign_type"])
			if !signType.IsValid() {
				return errors.New("Unknown sign type")
			}
		}
		if !signType.IsValid() {
			signType = options.SignType()
		}

		// 验证签名
		sign := SignMchXML(reqXML, signType, config.WechatMchKey())
		suppliedSign := reqXML["sign"]
		if suppliedSign == "" || suppliedSign != sign {
			return errors.New("Sign error")
		}

		// 通过了，执行 handler
		return handler(ctx, reqXML)

	}, options)

}
