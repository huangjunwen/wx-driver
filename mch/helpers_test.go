package mch

import (
	"github.com/stretchr/testify/assert"
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
