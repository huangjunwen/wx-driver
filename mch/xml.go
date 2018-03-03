package mch

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"hash"
	"sort"
)

// MchXML 代表微信支付接口的 xml 数据，形如：
//
//   <xml>
//     <fieldName1>fieldValue1</fieldName1>
//     <fieldName2>fieldValue2</fieldName2>
//     ...
//   </xml>
//
// 为一层 xml，且字段名唯一：这是由签名算法决定的，假如字段名不唯一，例如：
//
//   <xml>
//     <fieldName1>fieldValue1</fieldName1>
//     <fieldName1>fieldValue2</fieldName1>
//     ...
//   </xml>
//
// 那么签名算法该使用
//
//   fieldName1=fieldValue1&fieldName1=fieldValue2
//
// 还是
//
//   fieldName1=fieldValue2&fieldName1=fieldValue1
//
// 进行签名呢？因此可以使用 map
type MchXML map[string]string

// UnmarshalXML 实现 xml.Unmarshaler 接口
func (x MchXML) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if start.Name.Local != "xml" {
		return fmt.Errorf("Expect <xml> as top element but got <%s>", start.Name.Local)
	}
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		// <xml> 层
		switch t0 := t.(type) {
		default:
			// 忽略
		case xml.EndElement:
			// 结束
			return nil

		case xml.StartElement:
			// xml tag 为 fieldName，检查唯一性
			fieldName := t0.Name.Local
			if _, ok := x[fieldName]; ok {
				return fmt.Errorf("Duplicate field name <%s> in MchXML", fieldName)
			}

			// 取出 <xml><sub> 层的 chardata 作为 fieldValue
			// Unmarshal maps an XML element to a string or []byte by saving the concatenation of
			// that element's character data in the string or []byte. The saved []byte is never nil.
			var fieldValue string
			if err := d.DecodeElement(&fieldValue, &t0); err != nil {
				return err
			}

			x[fieldName] = fieldValue
		}
	}
	return nil
}

var (
	xmlStart = xml.StartElement{Name: xml.Name{Local: "xml"}}
	xmlEnd   = xml.EndElement{Name: xml.Name{Local: "xml"}}
)

// MarshalXML 实现 xml.Marshaler 接口
func (x MchXML) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if err := e.EncodeToken(xmlStart); err != nil {
		return err
	}
	for fieldName, fieldValue := range x {
		if err := e.EncodeToken(xml.StartElement{Name: xml.Name{Local: fieldName}}); err != nil {
			return err
		}
		if err := e.EncodeToken(xml.CharData(fieldValue)); err != nil {
			return err
		}
		if err := e.EncodeToken(xml.EndElement{Name: xml.Name{Local: fieldName}}); err != nil {
			return err
		}
	}
	if err := e.EncodeToken(xmlEnd); err != nil {
		return err
	}
	return nil
}

// Sign 对该 xml 数据进行签名，签名算法见微信支付《安全规范》，signType 为空时默认使用 MD5，
// x 中 sign 字段和空值字段皆不参与签名
func (x MchXML) Sign(signType SignType, key string) string {
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
