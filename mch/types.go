package mch

import (
	"fmt"
)

// MchResponse 为微信支付接口响应的公共部分，接口响应可以把该结构 embed 进去
type MchResponse struct {
	// 业务结果字段
	ResultCode string
	ErrCode    string
	ErrCodeDes string

	// 备查字段
	AppID    string
	MchID    string
	NonceStr string
	Sign     string
}

// IsSuccess 返回该响应的业务结果是否成功，若返回 false，可调用 Error 方法获得具体错误
func (resp *MchResponse) IsSuccess() bool {
	return resp.ResultCode == "SUCCESS"
}

// Error 当业务结果不成功的错误原因，若业务结果成功返回 nil
func (resp *MchResponse) Error() error {
	if resp.IsSuccess() {
		return nil
	}
	return fmt.Errorf("Mch result errcode=%+q errmsg=%+q", resp.ErrCode, resp.ErrCodeDes)
}

func (resp *MchResponse) mchXMLIter(i int, fieldName, fieldValue string) error {
	switch fieldName {
	case "result_code":
		resp.ResultCode = fieldValue
	case "err_code":
		resp.ErrCode = fieldValue
	case "err_code_des":
		resp.ErrCodeDes = fieldValue
	case "appid":
		resp.AppID = fieldValue
	case "mch_id":
		resp.MchID = fieldValue
	case "nonce_str":
		resp.NonceStr = fieldValue
	case "sign":
		resp.Sign = fieldValue
	}
	return nil
}
