package mch

import (
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestMchXMLUnmarshalXML(t *testing.T) {
	assert := assert.New(t)
	for _, testCase := range []struct {
		Src          string
		ExpectOK     bool
		ExpectResult map[string]string
	}{
		{"<xml></notxml>", false, nil},                                          // xml 错误
		{"<notxml></notxml>", false, nil},                                       // 顶层元素错误
		{"<xml></xml>", true, map[string]string{}},                              // 正确，没有字段
		{"<xml><a>1</a><a>2</a></xml>", false, nil},                             // 重复字段
		{"<xml><a></a></xml>", true, map[string]string{"a": ""}},                // 正确，空字段值
		{"<xml><a>b</a><c/></xml>", true, map[string]string{"a": "b", "c": ""}}, // 正确，自闭合元素
		{"<xml><a>b<c/>d</a></xml>", true, map[string]string{"a": "bd"}},        // 正确（其实也可以错误），深于一层的元素忽略掉
	} {

		x := MchXML(make(map[string]string))
		err := xml.Unmarshal([]byte(testCase.Src), &x)

		log.Printf("xml=%#v err=%#v\n", x, err)

		if testCase.ExpectOK {
			assert.NoErrorf(err, "Expect has no error")
		} else {
			assert.Errorf(err, "Expect has error")
		}

		if testCase.ExpectResult != nil {
			assert.Equalf(testCase.ExpectResult, map[string]string(x), "Result is not expect")
		}

	}
}

func TestMchXMLSign(t *testing.T) {
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
	assert.Equal("9A0A8659F005D6984697E2CA0A9CF3B7", x.Sign(SignTypeInvalid, key))
	assert.Equal("9A0A8659F005D6984697E2CA0A9CF3B7", x.Sign(SignTypeMD5, key))
	assert.Equal("6A9AE1657590FD6257D693A078E1C3E4BB6BA4DC30B23E0EE2496E54170DACD6", x.Sign(SignTypeHMACSHA256, key))
}
