package mch

import (
	"encoding/xml"
	"fmt"
	"sort"
)

// MchXML 代表微信支付接口中的 xml 数据，为一层的 xml，根节点为 <xml>:
// 	 <xml>
//     <field1>value1</field1>
//     <field2>value2</field2>
//     ...
//   </xml>
type MchXML struct {
	XMLName struct{}      `xml:"xml"`
	Fields  []MchXMLField `xml:",any"`
}

type MchXMLField struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
}

// Field 返回第 i 各字段的名字和值
func (x *MchXML) Field(i int) (fieldName, fieldValue string) {
	if i < 0 || i > len(x.Fields)-1 {
		return "", ""
	}
	field := x.Fields[i]
	return field.XMLName.Local, field.Text
}

// Len 返回字段数
func (x *MchXML) Len() int {
	return len(x.Fields)
}

// AddField 添加新的字段
func (x *MchXML) AddField(fieldName, fieldValue string) {
	if fieldValue == "" {
		return
	}
	x.Fields = append(x.Fields, MchXMLField{
		XMLName: xml.Name{Local: fieldName},
		Text:    fieldValue,
	})
}

// SortFields 对所有字段进行字典序的排序并检查字段名唯一性
func (x *MchXML) SortFields() {
	sort.Slice(x.Fields, func(i, j int) bool {
		return x.Fields[i].XMLName.Local < x.Fields[j].XMLName.Local
	})
	// 检查字段名唯一性
	prev := ""
	for i := 0; i < x.Len(); i++ {
		name, _ := x.Field(i)
		if name == prev {
			panic(fmt.Errorf("Duplicate mch xml field name %+q", name))
		}
		prev = name
	}
}
