package mch

import (
	"time"
	"strconv"
	"fmt"
)

type RefundStatus string

const (
	RefundStatusSUCCESS     RefundStatus = "SUCCESS"     // 退款成功
	RefundStatusREFUNDCLOSE RefundStatus = "REFUNDCLOSE" // 退款关闭
	RefundStatusPROCESSING  RefundStatus = "PROCESSING"   // 退款处理中
	RefundStatusCHANGE      RefundStatus = "CHANGE"       // 退款异常，退款到银行发现用户的卡作废或者冻结了，导致原路退款银行卡失败，可前往商户平台交易中心，手动处理此笔退款
)

type QueryRefundRequest struct {
	// 必选参数, 四选一
	TransactionId string // 微信订单号
	OutTradeNo    string // 商户订单号
	OutRefundNo   string // 商户退款单号
	RefundId      string // 微信退款单号
}

type QueryRefundResponse struct {
	// 必选返回
	TransactionId string // 微信订单号
	OutTradeNo    string // 商户系统内部的订单号

	RefundItems []RefundItem

	// 可选返回
	TotalFee int // 订单总金额，单位为分，只能为整数
	CashFee  int // 现金支付金额，单位为分，只能为整数
	SettlementTotalFee int    // 应结订单金额=订单金额-非充值代金券金额，应结订单金额<=订单金额
	FeeType            string // 订单金额货币类型，符合ISO 4217标准的三位字母代码，默认人民币：CNY
	CashFeeType        string // 货币类型，符合ISO 4217标准的三位字母代码，默认人民币：CNY
}

type RefundItem struct {
	OutRefundNo      string       // 商户退款单号
	RefundId         string       // 微信退款单号
	RefundFee        int          // 申请退款金额
	RefundStatus     RefundStatus // 退款状态

	// 可选返回
	RefundChannel       string    // 退款渠道
	SettlementRefundFee int       // 退款金额=申请退款金额-非充值代金券退款金额，退款金额<=申请退款金额
	RefundAccount       string    // 退款资金来源
	RefundRecvAccout string       // 退款入账账户
	RefundSuccessTime   time.Time // 退款成功时间
	// TODO: coupon list
}

func (client *Client) QueryRefund(req *QueryRefundRequest) (rep *QueryRefundResponse, err error) {
	reqMap := make(map[string]string)
	if req.TransactionId != "" {
		reqMap["transaction_id"] = req.TransactionId
	}
	if req.OutTradeNo != "" {
		reqMap["out_trade_no"] = req.OutTradeNo
	}
	if req.OutRefundNo != "" {
		reqMap["out_refund_no"] = req.OutRefundNo
	}
	if req.RefundId != "" {
		reqMap["refund_id"] = req.RefundId
	}

	repMap, err := client.PostXML("/pay/refundquery", reqMap)
	if err != nil {
		return nil, err
	}

	rep = &QueryRefundResponse{
		TransactionId: repMap["transaction_id"],
		OutTradeNo:    repMap["out_trade_no"],
		FeeType:       repMap["fee_type"],
		CashFeeType:   repMap["cash_fee_type"],
	}

	if str, ok := repMap["total_fee"]; ok {
		rep.TotalFee, err = strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
	}
	if str, ok := repMap["cash_fee"]; ok {
		rep.CashFee, err = strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
	}
	if str, ok := repMap["settlement_total_fee"]; ok {
		rep.SettlementTotalFee, err = strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
	}

	refundCount, err := strconv.Atoi(repMap["refund_count"])
	if err != nil {
		return nil, err
	}

	rep.RefundItems = make([]RefundItem, refundCount)
	for i := 0; i < refundCount; i++ {
		item := &rep.RefundItems[i]
		index := strconv.Itoa(i)
		item.OutRefundNo = repMap["out_refund_no_"+index]
		item.RefundId = repMap["refund_id_"+index]
		item.RefundStatus = RefundStatus(repMap["refund_status_"+index])
		item.RefundRecvAccout = repMap["refund_recv_accout_"+index]
		item.RefundChannel = repMap["refund_channel_"+index]
		item.RefundAccount = repMap["refund_account_"+index]

		item.RefundFee, err = strconv.Atoi(repMap["refund_fee_"+index])
		if err != nil {
			return nil, err
		}
		if str := repMap["settlement_refund_fee_"+index]; str != "" {
			item.SettlementRefundFee, err = strconv.Atoi(str)
			if err != nil {
				return nil, err
			}
		}
		if str := repMap["refund_success_time_"+index]; str != "" {
			item.RefundSuccessTime, err = time.ParseInLocation("2006-01-02 15:04:05", str, BeijingLocation)
			if err != nil {
				return nil, err
			}
		}
	}

	if req.TransactionId != "" && rep.TransactionId != "" && req.TransactionId != rep.TransactionId {
		err = fmt.Errorf("transaction_id mismatch, have: %s, want: %s", rep.TransactionId, req.TransactionId)
		return nil, err
	}
	if req.OutTradeNo != "" && rep.OutTradeNo != "" && req.OutTradeNo != rep.OutTradeNo {
		err = fmt.Errorf("out_trade_no mismatch, have: %s, want: %s", rep.OutTradeNo, req.OutTradeNo)
		return nil, err
	}

	return rep, err
}
