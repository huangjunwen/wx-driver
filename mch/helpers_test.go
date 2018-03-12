package mch

import (
	"bytes"
	"context"
	"encoding/xml"
	"github.com/huangjunwen/wxdriver/conf"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestClient []byte

func (c TestClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(c))),
	}, nil
}

var (
	// 微信支付安全规范 (https://pay.weixin.qq.com/wiki/doc/api/danpin.php?chapter=4_3) 上的测试用例
	config = conf.DefaultConfig{
		conf.WechatAppIDName:  "wxd930ea5d5a258f4f",
		conf.WechatMchIDName:  "10000100",
		conf.WechatMchKeyName: "192006250b4c09247ec02edce69f6a2d",
	}
)

func TestSignMchXML(t *testing.T) {
	assert := assert.New(t)
	// 微信支付安全规范 (https://pay.weixin.qq.com/wiki/doc/api/danpin.php?chapter=4_3) 上的测试用例
	x := MchXML{
		"appid":       config.WechatAppID(),
		"mch_id":      config.WechatMchID(),
		"device_info": "1000",
		"body":        "test",
		"nonce_str":   "ibuaiVcKdpRxkhJA",
	}
	key := config.WechatMchKey()

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

func TestPostMchXML(t *testing.T) {
	assert := assert.New(t)

	for _, testCase := range []struct {
		Data          string
		ExpectSuccess bool
	}{
		{
			"<xml></notxml>",
			false,
		}, // 非法 xml
		{
			"<xml></xml>",
			false,
		}, // 非成功 return_code
		{
			"<xml><return_code>XX</return_code></xml>",
			false,
		}, // 非成功 return_code
		{
			"<xml><return_code>FAIL</return_code></xml>",
			false,
		}, // 非成功 return_code
		{
			"<xml><return_code>SUCCESS</return_code></xml>",
			false,
		}, // 无签名
		{
			"<xml><return_code>SUCCESS</return_code><sign>2C2B2A1D626E750FCFD0ED661E80E3AB</sign></xml>",
			false,
		}, // 签名不一致
		{
			"<xml><return_code>SUCCESS</return_code><sign>2c2b2a1d626e750fcfd0ed661e80e3aa</sign></xml>",
			false,
		}, // 签名小写
		{
			`<xml>
			<return_code><![CDATA[SUCCESS]]></return_code>
			<sign>2C2B2A1D626E750FCFD0ED661E80E3AA</sign>
			</xml>`,
			false,
		}, // 缺少 result_code
		{
			`<xml>
			<appid><![CDATA[wxd930ea5d5a258f40]]></appid>
			<return_code><![CDATA[SUCCESS]]></return_code>
			<sign>82D7E4E9DD2AF081B9299A6D2662BC68</sign>
			</xml>`,
			false,
		}, // appid 不一致
		{
			`<xml>
			<appid><![CDATA[wxd930ea5d5a258f4f]]></appid>
			<mch_id><![CDATA[10000101]]></mch_id>
			<return_code><![CDATA[SUCCESS]]></return_code>
			<sign>94A7200F4A777691304A3095309184E7</sign>
			</xml>`,
			false,
		}, // mch_id 不一致
		{
			`<xml>
			<appid><![CDATA[wxd930ea5d5a258f4f]]></appid>
			<mch_id><![CDATA[10000100]]></mch_id>
			<result_code><![CDATA[FAIL]]></result_code>
			<return_code><![CDATA[SUCCESS]]></return_code>
			<sign>FE153FF7F03BE9A3A2C7188464320B1D</sign>
			</xml>`,
			false,
		}, // result_code 不是 SUCCESS
		{
			`<xml>
			<appid><![CDATA[wxd930ea5d5a258f4f]]></appid>
			<mch_id><![CDATA[10000100]]></mch_id>
			<result_code><![CDATA[SUCCESS]]></result_code>
			<return_code><![CDATA[SUCCESS]]></return_code>
			<sign>12918CCB221CC80C0961BCC5903F5B25</sign>
			</xml>`,
			true,
		}, // 通过
	} {
		client := TestClient([]byte(testCase.Data))

		_, err := postMchXML(context.Background(), config, "/", MchXML{}, MustOptions(
			UseClient(client),
		))

		if testCase.ExpectSuccess {
			assert.NoError(err, "Expect no error but got %+q, response data:\n\n%+q", err, testCase.Data)
		} else {
			assert.Error(err, "Expect error but got nil, response data:\n\n%+q", testCase.Data)
		}

	}

}

func TestHandleSignedMchXML(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(handleSignedMchXML(func(context.Context, MchXML) error {
		return nil
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

		assert.Equal(testCase.ExpectReturnCode, respXML["return_code"], "Expect return_code %+q but got %+q, request data: \n\n%+q", testCase.ExpectReturnCode, respXML["return_code"], testCase.Data)

		assert.Equal(testCase.ExpectReturnMsg, respXML["return_msg"], "Expect return_msg %+q but got %+q, request data: \n\n%+q", testCase.ExpectReturnMsg, respXML["return_msg"], testCase.Data)
	}

}
