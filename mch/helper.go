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
	"strings"
)

// signMchXML 对 mchXML 签名；返回 actual 和 supplied 两个签名：
// actual 是使用微信支付签名算法计算出来的签名
// supplied 则是从 mchXML 里直接提取的 sign 字段（若有），否则为空
func signMchXML(x *mchXML, signType SignType, key string) (actual, supplied string) {
	// 选择 hash
	var h hash.Hash
	switch signType {
	case SignTypeHMACSHA256:
		h = hmac.New(sha256.New, []byte(key))
	default:
		h = md5.New()
	}

	// 字典序排序并验证唯一性
	x.SortUniqueFields()

	// 开始签名
	x.EachField(func(i int, fieldName, fieldValue string) error {
		// 值为空不参与签名
		if fieldValue == "" {
			return nil
		}

		// sign 不参与签名，只是记录下来返回之，用于对比
		if fieldName == "sign" {
			// 转换成大写
			supplied = strings.ToUpper(fieldValue)
			return nil
		}

		h.Write([]byte(fieldName))
		h.Write([]byte("="))
		h.Write([]byte(fieldValue))
		h.Write([]byte("&"))
		return nil
	})
	h.Write([]byte("key="))
	h.Write([]byte(key))

	// 需要大写
	actual = fmt.Sprintf("%X", h.Sum(nil))
	return
}

func postMchXML(ctx context.Context, config Configuration, url string, input, output *mchXML, opts Options) error {
	// 签名方式默认为 MD5
	signType := opts.SignType
	if !signType.IsValid() {
		signType = SignTypeMD5
	}

	// http client 依次选择：opts.HTTPClient > DefaultOptions.HTTPClient > wxdriver.DefaultHTTPClient > http.DefaultClient
	client := opts.HTTPClient
	if client == nil {
		client = DefaultOptions.HTTPClient
	}
	if client == nil {
		client = wxdriver.DefaultHTTPClient
	}
	if client == nil {
		client = http.DefaultClient
	}

	// 添加一些公共字段
	input.AddField("appid", config.WechatAppID())
	input.AddField("mch_id", config.WechatPayMchID())
	input.AddField("sign_type", signType.String())
	input.AddField("nonce_str", wxdriver.NonceStr(16)) // 32 位以内

	// 计算签名，NOTE：同时检查字段唯一性
	actualSign, suppliedSign := signMchXML(input, signType, config.WechatPayKey())
	if suppliedSign != "" {
		// 输入不应该有 <sign>
		panic(fmt.Errorf("Request should not have <sign> but got one %+q", suppliedSign))
	}
	input.AddField("sign", actualSign)

	// 编码
	buf := &bytes.Buffer{}
	encoder := xml.NewEncoder(buf)
	if err := encoder.Encode(input); err != nil {
		return err
	}

	// 构造请求
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	// 调用!
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// 解码
	decoder := xml.NewDecoder(resp.Body)
	output.Reset()
	if err := decoder.Decode(output); err != nil {
		return err
	}

	// 获取 return_code 和 return_msg
	returnCode, returnMsg := "", ""
	output.EachField(func(_ int, name, val string) error {
		switch name {
		case "return_code":
			returnCode = val
		case "return_msg":
			returnMsg = val
		}
		return nil
	})
	if returnCode != "SUCCESS" {
		// return_code 不成功时没有签名，所以直接返回其错误信息
		return fmt.Errorf("Response return %+q with msg: %+q", returnCode, returnMsg)
	}

	// 验证签名，NOTE：同时验证字段唯一性
	actualSign, suppliedSign = signMchXML(output, signType, config.WechatPayKey())
	if actualSign != suppliedSign {
		return fmt.Errorf("Response has actual sign %+q... but got %+q", actualSign[:8], suppliedSign)
	}

	// 全部通过
	return nil

}
