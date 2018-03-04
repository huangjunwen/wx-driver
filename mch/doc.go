// Package mch 包含微信支付 sdk
//
// 各接口一般形如：
//
//   X(ctx context.Context, config Configuration, req *XRequest, options ...Option) (resp *XResponse, err error)
//
// 其中 X 是接口名称，各参数说明如下：
//
//   ctx 上下文
//   config 微信支付配置，此项是包含接口必须的参数，如应用 ID/商户 ID 等
//   req 请求
//   options 额外选项，如非默认的 http client，签名方式等
//   resp 响应
//   err 错误，只有当通讯成功且业务结果成功时，返回空 err
package mch
