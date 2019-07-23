// Copyright © 2018 Huang junwen
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"context"
	"errors"
	"log"

	//"github.com/davecgh/go-spew/spew"
	"github.com/huangjunwen/wxdriver/conf"
	"github.com/huangjunwen/wxdriver/mch"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	mchConfig *conf.DefaultConfig
	// mchCmd represents the mch command
	mchCmd = &cobra.Command{
		Use:     "mch",
		Aliases: []string{"pay"},
		Short:   "Wechat payment API utilities",
		Long:    ``,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// BUG: 这里得手动调用 parent 的 PersistentPreRun，因为只会调用最接近叶节点的 PersistentPreRun
			// 见 https://github.com/spf13/cobra/issues/252
			if rootCmd.PersistentPreRunE != nil {
				if err := rootCmd.PersistentPreRunE(cmd, args); err != nil {
					return err
				}
			}
			return initMchConfig()
		},
	}
)

var (
	orderQueryRequest = &mch.OrderQueryRequest{}
	orderQueryCmd     = &cobra.Command{
		Use:   "orderquery",
		Short: "Query payment order",
		Long: `
Query payment order API: https://api.mch.weixin.qq.com/pay/orderquery

One of "transaction_id"/"out_trade_no" is required
		`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := mch.OrderQuery(context.Background(), mchConfig, orderQueryRequest)
			if err != nil {
				log.Fatal(err)
			}
			dumpJSON(map[string]interface{}{
				"raw": resp.MchXML,
			})
		},
	}
)

var (
	refundRequest = &mch.RefundRequest{}
	refundCmd     = &cobra.Command{
		Use:   "refund",
		Short: "Refund order",
		Long: `
Refund order API: https://api.mch.weixin.qq.com/secapi/pay/refund

One of "out_trade_no"/"out_refund_no"/"total_fee"/"refund_fee" is required
		`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := mch.Refund(
				context.Background(),
				mchConfig,
				refundRequest,
			)
			if err != nil {
				log.Fatal(err)
			}
			dumpJSON(map[string]interface{}{
				"raw": resp.MchXML,
			})
		},
	}
)

var (
	refundQueryRequest = &mch.RefundQueryRequest{}
	refundQueryCmd     = &cobra.Command{
		Use:   "refundquery",
		Short: "Query refund order",
		Long: `
Query refund order API: https://api.mch.weixin.qq.com/pay/refundquery

One of "transaction_id"/"out_trade_no"/"refund_id"/"out_refund_no" is required
		`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := mch.RefundQuery(context.Background(), mchConfig, refundQueryRequest)
			if err != nil {
				log.Fatal(err)
			}
			dumpJSON(map[string]interface{}{
				"raw": resp.MchXML,
			})
		},
	}
)

var (
	closeOrderRequest = &mch.CloseOrderRequest{}
	closeOrderCmd     = &cobra.Command{
		Use:   "closeorder",
		Short: "Close order",
		Long: `
Close order API: https://api.mch.weixin.qq.com/pay/closeorder

"out_trade_no" is required
		`,
		Run: func(cmd *cobra.Command, args []string) {
			err := mch.CloseOrder(context.Background(), mchConfig, closeOrderRequest)
			if err != nil {
				log.Fatal(err)
			}
			log.Print("close order ok")
		},
	}
)

func init() {
	// ----- mchCmd -----
	rootCmd.AddCommand(mchCmd)
	// PersistentFlags 是用于该命令以及所有其子命令的
	mchCmd.PersistentFlags().StringP("mch_id", "m", "", "Wechat payment mch id [required|configurable]")
	viper.BindPFlag("mch_id", mchCmd.PersistentFlags().Lookup("mch_id"))

	mchCmd.PersistentFlags().StringP("mch_key", "k", "", "Wechat payment mch key [required|configurable]")
	viper.BindPFlag("mch_key", mchCmd.PersistentFlags().Lookup("mch_key"))

	// ----- orderQueryCmd -----
	mchCmd.AddCommand(orderQueryCmd)
	orderQueryCmd.Flags().StringVar(&orderQueryRequest.TransactionID, "transaction_id", "", "Wechat payment transaction_id")
	orderQueryCmd.Flags().StringVar(&orderQueryRequest.OutTradeNo, "out_trade_no", "", "Wechat payment out_trade_no")

	// ----- refundCmd -----
	mchCmd.AddCommand(refundCmd)
	refundCmd.Flags().StringVar(&refundRequest.OutTradeNo, "out_trade_no", "", "Wechat payment out_trade_no")
	refundCmd.Flags().StringVar(&refundRequest.OutRefundNo, "out_refund_no", "", "Wechat payment out_refund_no")
	refundCmd.Flags().Uint64Var(&refundRequest.TotalFee, "total_fee", 0, "Wechat payment total_fee")
	refundCmd.Flags().Uint64Var(&refundRequest.RefundFee, "refund_fee", 0, "Wechat payment refund_fee")

	// ----- refundQueryCmd -----
	mchCmd.AddCommand(refundQueryCmd)
	refundQueryCmd.Flags().StringVar(&refundQueryRequest.TransactionID, "transaction_id", "", "Wechat payment transaction_id")
	refundQueryCmd.Flags().StringVar(&refundQueryRequest.OutTradeNo, "out_trade_no", "", "Wechat payment out_trade_no")
	refundQueryCmd.Flags().StringVar(&refundQueryRequest.RefundID, "refund_id", "", "Wechat payment refund_id")
	refundQueryCmd.Flags().StringVar(&refundQueryRequest.OutRefundNo, "out_refund_no", "", "Wechat payment out_refund_no")
	refundQueryCmd.Flags().UintVar(&refundQueryRequest.Offset, "offset", 0, "Offset of refund orders")

	// ----- closeOrderCmd -----
	mchCmd.AddCommand(closeOrderCmd)
	closeOrderCmd.Flags().StringVar(&closeOrderRequest.OutTradeNo, "out_trade_no", "", "Wechat payment out_trade_no")

}

func initMchConfig() error {
	appID := viper.GetString("appid")
	if appID == "" {
		return errors.New("'appid' is required but not found")
	}
	mchID := viper.GetString("mch_id")
	if mchID == "" {
		return errors.New("'mch_id' is required but not found")
	}
	mchKey := viper.GetString("mch_key")
	if mchKey == "" {
		return errors.New("'mch_key' is required but not found")
	}

	mchConfig = &conf.DefaultConfig{
		AppID:  appID,
		MchID:  mchID,
		MchKey: mchKey,
	}
	return nil
}
