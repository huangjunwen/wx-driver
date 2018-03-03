package mch

import (
	"encoding/xml"
	"sort"
)

// mchXML 代表微信支付接口中的 xml 数据，为一层的 xml
//
// 	 <xml>
//     <field1>value1</field1>
//     <field2>value2</field2>
//     ...
//   </xml>
//
type mchXML struct {
	XMLName struct{}      `xml:"xml"`
	Fields  []mchXMLField `xml:",any"`
	sorted  bool
}

type mchXMLField struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
}

// IterateFields 迭代 fields
func (x *mchXML) IterateFields(fns ...func(i int, fieldName, fieldValue string) error) error {
	for idx, field := range x.Fields {
		for _, fn := range fns {
			if err := fn(idx, field.XMLName.Local, field.Text); err != nil {
				return err
			}
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
	x.sorted = false
}

// SortFields 按字段名字典序排序
func (x *mchXML) SortFields() {
	if x.sorted {
		return
	}
	sort.Slice(x.Fields, func(a, b int) bool {
		return x.Fields[a].XMLName.Local < x.Fields[b].XMLName.Local
	})
	x.sorted = true
}

// CheckFieldsUniqueness 检查字段名的唯一性，若字段皆唯一，返回空串，否则返回第一个重复的字段；
// 该方法会调用 SortFields
func (x *mchXML) CheckFieldsUniqueness() (dupFieldName string) {
	x.SortFields()
	prevFieldName := ""
	for _, field := range x.Fields {
		fieldName := field.XMLName.Local
		if fieldName == prevFieldName {
			return fieldName
		}
		prevFieldName = fieldName
	}
	return ""
}
