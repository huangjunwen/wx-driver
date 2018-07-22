package mch

import (
	"testing"
	"time"

	"github.com/huangjunwen/wxdriver/conf"
	"github.com/huangjunwen/wxdriver/utils"
	"github.com/stretchr/testify/assert"
)

func TestJSReq(t *testing.T) {
	assert := assert.New(t)

	// https://pay.weixin.qq.com/wiki/doc/api/wxa/wxa_api.php?chapter=7_7&index=5 的例子
	config := &conf.DefaultConfig{
		AppID:  "wxd678efh567hg6787",
		MchKey: "qazwsxedcrfvtgbyhnujmikolp111111",
	}
	prepayID := "wx2017033010242291fcfe0db70013231072"

	nonceStr := utils.NonceStr
	now := utils.Now
	utils.NonceStr = func(n int) string {
		return "5K8264ILTKCH16CQ2502SI8ZNMTM67VS"
	}
	utils.Now = func() time.Time {
		return time.Unix(1490840662, 0)
	}
	defer func() {
		utils.NonceStr = nonceStr
		utils.Now = now
	}()

	assert.Equal("22D9B4E54AB1950F51E0649E8810ACD6",
		JSReqEx(config, prepayID, SignTypeMD5)["paySign"],
	)
}
