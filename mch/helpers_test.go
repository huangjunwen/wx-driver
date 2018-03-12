package mch

import (
	"bytes"
	"context"
	"encoding/xml"
	"github.com/huangjunwen/wxdriver/conf"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignMchXML(t *testing.T) {
	assert := assert.New(t)
	// 微信支付安全规范 (https://pay.weixin.qq.com/wiki/doc/api/danpin.php?chapter=4_3) 上的测试用例
	x := MchXML{
		"appid":       "wxd930ea5d5a258f4f",
		"mch_id":      "10000100",
		"device_info": "1000",
		"body":        "test",
		"nonce_str":   "ibuaiVcKdpRxkhJA",
	}
	key := "192006250b4c09247ec02edce69f6a2d"

	for _, testCase := range []struct {
		X        MchXML
		SignType SignType
		Key      string
		Expect   string
	}{
		{x, SignTypeInvalid, key, "9A0A8659F005D6984697E2CA0A9CF3B7"},
		{x, SignTypeMD5, key, "9A0A8659F005D6984697E2CA0A9CF3B7"},
		{x, SignTypeHMACSHA256, key, "6A9AE1657590FD6257D693A078E1C3E4BB6BA4DC30B23E0EE2496E54170DACD6"},
	} {
		assert.Equal(testCase.Expect, signMchXML(testCase.X, testCase.SignType, testCase.Key))
	}
}

func TestHandleSignedMchXML(t *testing.T) {
	assert := assert.New(t)
	// 微信支付安全规范 (https://pay.weixin.qq.com/wiki/doc/api/danpin.php?chapter=4_3) 上的测试用例
	config := conf.DefaultConfig{
		conf.WechatAppIDName:  "wxd930ea5d5a258f4f",
		conf.WechatMchIDName:  "10000100",
		conf.WechatMchKeyName: "192006250b4c09247ec02edce69f6a2d",
	}

	var retError error = nil
	ts := httptest.NewServer(handleSignedMchXML(func(context.Context, MchXML) error {
		return retError
	}, config, nil))
	defer ts.Close()

	for _, testCase := range []struct {
		Data             string
		ExpectReturnCode string
		ExpectReturnMsg  string
	}{
		{
			"<xml></notxml>",
			"FAIL",
			"Invalid xml",
		}, // 非法 xml 数据
		{
			"<xml></xml>",
			"FAIL",
			"Failed return_code",
		}, // 非成功 return_code
		{
			"<xml><return_code>XX</return_code></xml>",
			"FAIL",
			"Failed return_code",
		}, // 非成功 return_code
		{
			"<xml><return_code>FAIL</return_code></xml>",
			"FAIL",
			"Failed return_code",
		}, // 非成功 return_code
		{
			"<xml><return_code>SUCCESS</return_code></xml>",
			"FAIL",
			"Unknown app or mch",
		}, // 缺少 appid 和 mch_id
		{
			"<xml><return_code>SUCCESS</return_code><appid>wxd930ea5d5a258f4f</appid></xml>",
			"FAIL",
			"Unknown app or mch",
		}, // 缺少 mch_id
		{
			"<xml><return_code>SUCCESS</return_code><mch_id>10000100</mch_id></xml>",
			"FAIL",
			"Unknown app or mch",
		}, // 缺少 appid
		{
			"<xml><return_code>SUCCESS</return_code><appid>fakewxd930ea5d5a258f4f</appid><mch_id>10000100</mch_id></xml>",
			"FAIL",
			"Unknown app or mch",
		}, // appid 不一致
		{
			"<xml><return_code>SUCCESS</return_code><appid>wxd930ea5d5a258f4f</appid><mch_id>fake10000100</mch_id></xml>",
			"FAIL",
			"Unknown app or mch",
		}, // mch_id 不一致
		{
			"<xml><return_code>SUCCESS</return_code><appid>wxd930ea5d5a258f4f</appid><mch_id>10000100</mch_id><sign_type>XXX</sign_type></xml>",
			"FAIL",
			"Unknown sign type",
		}, // 未知 sign_type
		{
			"<xml><return_code>SUCCESS</return_code><appid>wxd930ea5d5a258f4f</appid><mch_id>10000100</mch_id></xml>",
			"FAIL",
			"Sign error",
		}, // 没有签名
		{
			"<xml><return_code>SUCCESS</return_code><appid>wxd930ea5d5a258f4f</appid><mch_id>10000100</mch_id><sign>4D70B2071E6998EF23F3415D7BE3AC14</sign></xml>",
			"FAIL",
			"Sign error",
		}, // 修改了签名
		{
			"<xml><return_code>SUCCESS</return_code><appid>wxd930ea5d5a258f4f</appid><mch_id>10000100</mch_id><sign>4D70B2071E6998EF23F3415D7BE3AC15</sign></xml>",
			"SUCCESS",
			"",
		}, // 使用微信支付接口签名校验工具计算的结果
		{
			"<xml><return_code>SUCCESS</return_code><appid>wxd930ea5d5a258f4f</appid><mch_id>10000100</mch_id><sign>4d70b2071e6998ef23f3415d7be3ac15</sign></xml>",
			"FAIL",
			"Sign error",
		}, // 签名必须是小写
	} {
		resp, err := http.Post(ts.URL, "application/xml", bytes.NewBufferString(testCase.Data))
		assert.NoError(err, "http.Post should have no error")

		respXML := MchXML{}
		err = xml.NewDecoder(resp.Body).Decode(&respXML)
		assert.NoError(err, "xml.Decoder.Decode should have no error")

		assert.Equal(testCase.ExpectReturnCode, respXML["return_code"], "Expect return_code %+q but got %+q, origin data: \n\n%+q", testCase.ExpectReturnCode, respXML["return_code"], testCase.Data)

		assert.Equal(testCase.ExpectReturnMsg, respXML["return_msg"], "Expect return_msg %+q but got %+q, origin data: \n\n%+q", testCase.ExpectReturnMsg, respXML["return_msg"], testCase.Data)
	}

}
