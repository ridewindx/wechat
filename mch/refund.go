package mch

import (
	"strconv"
	"fmt"
)

type RefundRequest struct {
	// TransactionId 和 OutTradeNo 二选一即可
	TransactionId string // 微信生成的订单号，在支付通知中有返回
	OutTradeNo    string // 商户侧传给微信的订单号

	// 必选参数
	OutRefundNo string // 商户系统内部的退款单号，商户系统内部唯一，同一退款单号多次请求只退一笔
	TotalFee    int    // 订单总金额，单位为分，只能为整数
	RefundFee   int    // 退款总金额，单位为分，只能为整数

	// 可选参数
	RefundFeeType string // 货币类型，符合ISO 4217标准的三位字母代码，默认人民币：CNY
	RefundDesc    string // 若商户传入，会在下发给用户的退款消息中体现退款原因
}

type RefundResponse struct {
	// 必选返回
	TransactionId string // 微信订单号
	OutTradeNo    string // 商户系统内部的订单号
	OutRefundNo   string // 商户退款单号
	RefundId      string // 微信退款单号

	RefundFee int // 退款总金额，单位为分，可以做部分退款
	TotalFee  int // 订单总金额，单位为分，只能为整数
	CashFee   int // 现金支付金额，单位为分，只能为整数

	// 可选返回
	SettlementRefundFee int    // 退款金额=申请退款金额-非充值代金券退款金额，退款金额<=申请退款金额
	SettlementTotalFee  int    // 应结订单金额=订单金额-非充值代金券金额，应结订单金额<=订单金额
	FeeType             string // 订单金额货币类型，符合ISO 4217标准的三位字母代码，默认人民币：CNY
	CashFeeType         string // 货币类型，符合ISO 4217标准的三位字母代码，默认人民币：CNY
	CashRefundFee       int    // 现金退款金额，单位为分，只能为整数
	CouponRefundFee     int    // 代金券退款金额<=退款金额，退款金额-代金券或立减优惠退款金额为现金
	CouponRefundCount   int    // 退款代金券使用数量
	// TODO: coupon list
}

func (client *Client) Refund(req *RefundRequest) (rep *RefundResponse, err error) {
	reqMap := make(map[string]string)
	if req.TransactionId != "" {
		reqMap["transaction_id"] = req.TransactionId
	}
	if req.OutTradeNo != "" {
		reqMap["out_trade_no"] = req.OutTradeNo
	}
	reqMap["out_refund_no"] = req.OutRefundNo
	reqMap["total_fee"] = strconv.Itoa(req.TotalFee)
	reqMap["refund_fee"] = strconv.Itoa(req.RefundFee)
	if req.RefundFeeType != "" {
		reqMap["refund_fee_type"] = req.RefundFeeType
	}
	if req.RefundDesc != "" {
		reqMap["refund_desc"] = req.RefundDesc
	}

	repMap, err := client.PostXMLWithCert("/secapi/pay/refund", reqMap)

	rep = &RefundResponse{
		TransactionId: repMap["transaction_id"],
		OutTradeNo:    repMap["out_trade_no"],
		OutRefundNo:   repMap["out_refund_no"],
		RefundId:      repMap["refund_id"],
		FeeType:       repMap["fee_type"],
		CashFeeType:   repMap["cash_fee_type"],
	}

	rep.RefundFee, err = strconv.Atoi(repMap["refund_fee"])
	if err != nil {
		return nil, err
	}
	if totalFee, ok := repMap["total_fee"]; ok {
		rep.TotalFee, err = strconv.Atoi(totalFee)
		if err != nil {
			return nil, err
		}
	}
	if cashFee, ok := repMap["cash_fee"]; ok {
		rep.CashFee, err = strconv.Atoi(cashFee)
		if err != nil {
			return nil, err
		}
	}
	if str := repMap["settlement_refund_fee"]; str != "" {
		rep.SettlementRefundFee, err = strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
	}
	if str := repMap["settlement_total_fee"]; str != "" {
		rep.SettlementTotalFee, err = strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
	}
	if str := repMap["cash_refund_fee"]; str != "" {
		rep.CashRefundFee, err = strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
	}
	if str := repMap["coupon_refund_fee"]; str != "" {
		rep.CouponRefundFee, err = strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
	}
	if str := repMap["coupon_refund_count"]; str != "" {
		rep.CouponRefundCount, err = strconv.Atoi(str)
		if err != nil {
			return nil, err
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
	if req.OutRefundNo != "" && rep.OutRefundNo != "" && req.OutRefundNo != rep.OutRefundNo {
		err = fmt.Errorf("out_refund_no mismatch, have: %s, want: %s", rep.OutRefundNo, req.OutRefundNo)
		return nil, err
	}
	if req.TotalFee != rep.TotalFee {
		err = fmt.Errorf("total_fee mismatch, have: %d, want: %d", rep.TotalFee, req.TotalFee)
		return nil, err
	}
	if req.RefundFee != rep.RefundFee {
		err = fmt.Errorf("refund_fee mismatch, have: %d, want: %d", rep.RefundFee, req.RefundFee)
		return nil, err
	}

	return rep, nil
}
