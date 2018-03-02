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
// 若字段非唯一返回错误，这是因为，假如有这样的 xml：
//
//   <xml>
//     <a>x</a>
//     <a>y</a>
//   </xml>
//
// 则无法确认是使用 'a=x&a=y' 还是 'a=y&a=x' 进行签名，两者都是合法的排序
func signMchXML(x *mchXML, signType SignType, key string) (actual, supplied string, err error) {
	// 选择 hash
	var h hash.Hash
	switch signType {
	case SignTypeHMACSHA256:
		h = hmac.New(sha256.New, []byte(key))
	default:
		h = md5.New()
	}

	// 字典序排序
	x.SortFields()

	// 检查唯一性
	dupFieldName := x.CheckFieldsUniqueness()
	if dupFieldName != "" {
		err = fmt.Errorf("Can't sign since duplicate field %+q", dupFieldName)
		return
	}

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

// postMchXML 调用 mch xml 接口，大致过程如下：
//   - 添加公共字段
//     - appid
//     - mch_id
//     - sign_type
//     - nonce_str
//   - 签名并添加 sign 字段
//   - 调用 api，等待结果或错误
//   - 验证通讯结果
//   - 验证签名
// 所以若返回 err 为 nil，表明上述所有过程均无出错，但业务上的结果需要调用者检查 output 各字段方可知道
func postMchXML(ctx context.Context, config Configuration, url string, reqXML, respXML *mchXML, opts *Options) error {
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

	// 签名方式默认为 MD5
	signType := opts.SignType
	if !signType.IsValid() {
		signType = SignTypeMD5
	}

	// 添加一些公共字段
	reqXML.AddField("appid", config.WechatAppID())
	reqXML.AddField("mch_id", config.WechatPayMchID())
	reqXML.AddField("sign_type", signType.String())
	reqXML.AddField("nonce_str", wxdriver.NonceStr(16)) // 32 位以内

	// 计算签名，同时验证字段唯一性
	actualSign, suppliedSign, err := signMchXML(reqXML, signType, config.WechatPayKey())
	if err != nil {
		return err
	}
	if suppliedSign != "" {
		return fmt.Errorf("Request should not have <sign> but got one %+q", suppliedSign)
	}
	reqXML.AddField("sign", actualSign)

	// 编码
	buf := &bytes.Buffer{}
	encoder := xml.NewEncoder(buf)
	if err := encoder.Encode(reqXML); err != nil {
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
	if err := decoder.Decode(respXML); err != nil {
		return err
	}

	// 提取 return_code/return_msg/appid/mch_id
	returnCode := ""
	returnMsg := ""
	appID := ""
	mchID := ""
	respXML.EachField(func(_ int, name, val string) error {
		switch name {
		case "return_code":
			returnCode = val
		case "return_msg":
			returnMsg = val
		case "appid":
			appID = val
		case "mch_id":
			mchID = val
		}
		return nil
	})

	// return_code 不成功时没有签名，所以直接返回其错误信息
	if returnCode != "SUCCESS" {
		return fmt.Errorf("Response return %+q with msg: %+q", returnCode, returnMsg)
	}

	// 验证签名，同时验证字段唯一性
	actualSign, suppliedSign, err = signMchXML(respXML, signType, config.WechatPayKey())
	if err != nil {
		return err
	}
	if actualSign != suppliedSign {
		return fmt.Errorf("Response has actual sign %+q but got %+q", actualSign, suppliedSign)
	}

	// 验证 appID 和 mchID
	if appID != "" && appID != config.WechatAppID() {
		return fmt.Errorf("Response appID expect %+q but got %+q", config.WechatAppID(), appID)
	}
	if mchID != "" && mchID != config.WechatPayMchID() {
		return fmt.Errorf("Response mchID expect %+q but got %+q", config.WechatPayMchID(), mchID)
	}

	// 全部通过
	return nil

}
