package mch

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"time"
)

// MchXML 代表微信支付接口的 xml 数据，形如：
//
//   <xml>
//     <fieldName1>fieldValue1</fieldName1>
//     <fieldName2>fieldValue2</fieldName2>
//     ...
//   </xml>
//
// 为一层 xml，字段名唯一：这是由签名算法决定的，假如字段名不唯一，例如：
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

func (x MchXML) extractString(target *string, fieldName string, err *error) {
	if *err != nil {
		return
	}
	*target = x[fieldName]
}

func (x MchXML) extractTime(target *time.Time, fieldName string, layout string, err *error) {
	if *err != nil {
		return
	}
	if fieldValue := x[fieldName]; fieldValue != "" {
		*target, *err = time.Parse(layout, fieldValue)
	}
}

func (x MchXML) extractTimeCompact(target *time.Time, fieldName string, err *error) {
	x.extractTime(target, fieldName, "20060102150405", err)
}

func (x MchXML) extractUint64(target *uint64, fieldName string, err *error) {
	if *err != nil {
		return
	}
	if fieldValue := x[fieldName]; fieldValue != "" {
		*target, *err = strconv.ParseUint(fieldValue, 10, 64)
	}
}

func (x MchXML) extractInt64(target *int64, fieldName string, err *error) {
	if *err != nil {
		return
	}
	if fieldValue := x[fieldName]; fieldValue != "" {
		*target, *err = strconv.ParseInt(fieldValue, 10, 64)
	}
}

func (x MchXML) extractTradeType(target *TradeType, fieldName string, err *error) {
	if *err != nil {
		return
	}
	if fieldValue := x[fieldName]; fieldValue != "" {
		*target = ParseTradeType(fieldValue)
		if !target.IsValid() {
			*err = fmt.Errorf("Unsupported trade type %+q", fieldValue)
		}
	}
}

func (x MchXML) extractTradeState(target *TradeState, fieldName string, err *error) {
	if *err != nil {
		return
	}
	if fieldValue := x[fieldName]; fieldValue != "" {
		*target = ParseTradeState(fieldValue)
		if !target.IsValid() {
			*err = fmt.Errorf("Unsupported trade state %+q", fieldValue)
		}
	}
}

func (x MchXML) extractSignType(target *SignType, fieldName string, err *error) {
	if *err != nil {
		return
	}
	if fieldValue := x[fieldName]; fieldValue != "" {
		*target = ParseSignType(fieldValue)
		if !target.IsValid() {
			*err = fmt.Errorf("Unsupported sign type %+q", fieldValue)
		}
	}
}

func (x MchXML) extractRefundStatus(target *RefundStatus, fieldName string, err *error) {
	if *err != nil {
		return
	}
	if fieldValue := x[fieldName]; fieldValue != "" {
		*target = ParseRefundStatus(fieldValue)
		if !target.IsValid() {
			*err = fmt.Errorf("Unsupported refund status %+q", fieldValue)
		}
	}
}

func (x MchXML) fillString(src string, fieldName string) {
	x[fieldName] = src
}

func (x MchXML) fillTime(src time.Time, fieldName string, layout string) {
	x[fieldName] = src.Format(layout)
}

func (x MchXML) fillTimeCompact(src time.Time, fieldName string) {
	x.fillTime(src, fieldName, "20060102150405")
}

func (x MchXML) fillUint64(src uint64, fieldName string) {
	x[fieldName] = strconv.FormatUint(src, 10)
}

func (x MchXML) fillInt64(src int64, fieldName string) {
	x[fieldName] = strconv.FormatInt(src, 10)
}

func (x MchXML) fillStringer(stringer fmt.Stringer, fieldName string) {
	x[fieldName] = stringer.String()
}
