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
		ExpectResult map[string][]byte
	}{
		{"<xml></notxml>", false, nil},                                                        // xml 错误
		{"<notxml></notxml>", false, nil},                                                     // 顶层元素错误
		{"<xml></xml>", true, map[string][]byte{}},                                            // 正确，没有字段
		{"<xml><a>1</a><a>2</a></xml>", false, nil},                                           // 重复字段
		{"<xml><a></a></xml>", true, map[string][]byte{"a": []byte{}}},                        // 正确，空字段值
		{"<xml><a>b</a><c/></xml>", true, map[string][]byte{"a": []byte("b"), "c": []byte{}}}, // 正确，自闭合元素
		{"<xml><a>b<c/>d</a></xml>", true, map[string][]byte{"a": []byte("bd")}},              // 正确（其实也可以错误），深于一层的元素忽略掉
	} {

		x := MchXML(make(map[string][]byte))
		err := xml.Unmarshal([]byte(testCase.Src), &x)

		log.Printf("xml=%#v err=%#v\n", x, err)

		if testCase.ExpectOK {
			assert.NoErrorf(err, "Expect has no error")
		} else {
			assert.Errorf(err, "Expect has error")
		}

		if testCase.ExpectResult != nil {
			assert.Equalf(testCase.ExpectResult, map[string][]byte(x), "Result is not expect")
		}

	}
}
