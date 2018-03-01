package mch

import (
	"encoding/xml"
	"fmt"
	"sort"
)

// mchXML 代表微信支付接口中的 xml 数据，为一层的 xml，根节点为 <xml>:
//
// 	 <xml>
//     <field1>value1</field1>
//     <field2>value2</field2>
//     ...
//   </xml>
//
// 字段名应该唯一
type mchXML struct {
	XMLName struct{}      `xml:"xml"`
	Fields  []mchXMLField `xml:",any"`
	// 自带一个 buf 给 Fields 使用，可减少一次 alloc
	fieldsBuf [32]mchXMLField
}

type mchXMLField struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
}

// Reset 清空 mchXML
func (x *mchXML) Reset() {
	x.Fields = x.fieldsBuf[:0]
}

// EachField 迭代 fields
func (x *mchXML) EachField(fn func(i int, fieldName, fieldValue string) error) error {
	for idx, field := range x.Fields {
		if err := fn(idx, field.XMLName.Local, field.Text); err != nil {
			return err
		}
	}
	return nil
}

// AddField 添加新的字段
func (x *mchXML) AddField(fieldName, fieldValue string) {
	x.Fields = append(x.Fields, mchXMLField{
		XMLName: xml.Name{Local: fieldName},
		Text:    fieldValue,
	})
}

// SortUniqueFields 按字段名字典序排序并检查字段名唯一性，若不唯一则 panic
func (x *mchXML) SortUniqueFields() {
	// 排序
	sort.Slice(x.Fields, func(a, b int) bool {
		return x.Fields[a].XMLName.Local < x.Fields[b].XMLName.Local
	})

	// 检查唯一性
	prevFieldName := ""
	for _, field := range x.Fields {
		fieldName := field.XMLName.Local
		if fieldName == prevFieldName {
			panic(fmt.Errorf("Duplicate mch xml field name %+q", prevFieldName))
		}
		prevFieldName = fieldName
	}
}
