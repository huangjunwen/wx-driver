package mch

import (
	"encoding/xml"
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
}

type mchXMLField struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
}

// ForeachField 迭代 fields
func (x *mchXML) ForeachField(fn func(i int, fieldName, fieldValue string) error) error {
	for idx, field := range x.Fields {
		if err := fn(idx, field.XMLName.Local, field.Text); err != nil {
			return err
		}
	}
	return nil
}

// AddField 添加新的字段
func (x *mchXML) AddField(fieldName, fieldValue string) {
	if fieldValue == "" {
		return
	}
	x.Fields = append(x.Fields, mchXMLField{
		XMLName: xml.Name{Local: fieldName},
		Text:    fieldValue,
	})
}

// SortFields 对所有字段按字段名进行字典序排序
func (x *mchXML) SortFields() {
	sort.Slice(x.Fields, func(i, j int) bool {
		return x.Fields[i].XMLName.Local < x.Fields[j].XMLName.Local
	})
}
