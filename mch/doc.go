// Package mch 包含微信支付 sdk.
//
// ##### 接口形式 ########################################################
//
// 各接口一般形如：
//
//   X(ctx context.Context, config Config, req *XRequest, opts ...Option) (resp *XResponse, err error)
//
// 其中 X 是接口名称，各参数说明如下：
//
//   ctx 上下文
//   config 微信支付配置，此项是包含接口必须的参数，如应用 ID/商户 ID 等
//   req 请求
//   opts 额外选项，如非默认的 http client，签名方式等
//   resp 响应
//   err 错误，只有当通讯成功且业务结果成功时，返回空 err
//
// ##### https ########################################################
//
// 有些接口（例如退款接口）需要客户端证书方可方可调用，最简便的方法是在 DefaultOptions
// 中统一使用带客户端证书的 http client：
//
//  import (
//  	"github.com/huangjunwen/wxdriver/mch"
//  	"github.com/huangjunwen/wxdriver/utils"
//  	"time"
//  )
//
//  mch.DefaultOptions = mch.MustOptions(
//  	mch.UseClient(&http.Client{
//  		Transport: &http.Transport{
//  			TLSClientConfig: utils.MustTLSConfig(certPEM, keyPEM, nil),
//  		},
//  		Timeout: 30 * time.Second,
//  	}),
//  )
//
// ##### 金额说明 ########################################################
//
// 接口里有各种金额，如：
//
//   - 标价金额 total_fee / 标价货币 fee_type
//   - 现金支付金额 cash_fee / 现金支付货币 cash_fee_type
//   - 退款金额 refund_fee / 退款货币 refund_fee_type
//   - 现金退款金额 cash_refund_fee / 现金退款货币 cash_refund_fee_type
//   - 优惠金额
//
// 标价货币和退款货币必须一致，现金支付货币跟现金退款货币也是一致的（即用户实际使用什么货币支付，则接受什么样的货币退款），并且满足：
//
//  refund_fee/total_fee == cash_refund_fee/cash_fee
//
// 即退款金额和标价金额确定退款比例，然后按此比例乘以用户实际现金支付金额得出现金退款金额
//
// 例如：
//
//   - 标价金额 1 美元
//   - 优惠金额 0.5 美元
//   - 现金支付 3.5 人民币 （== 0.5 美元，按汇率 7 来算）
//
// 则若退款：
//
//   - 退款金额 0.6 美元 （退 60%）
//   - 现金退款金额 2.1 人民币（== 3.5 人民币 * 0.6)
package mch
