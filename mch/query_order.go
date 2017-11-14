package mch

import (
	"time"
	"strconv"
	"fmt"
)

type QueryOrderRequest struct {
	TransactionID string // 微信的订单号，优先使用
	OutTradeNo    string // 商户系统内部的订单号，当没提供transaction_id时需要传这个
}

type OrderInfo struct {
	DeviceInfo     string // 终端设备号
	OpenID         string // 用户在商户appid下的唯一标识
	IsSubscribe    bool   // 用户是否关注公众账号，仅在公众账号类型支付有效
	SubOpenID      string // 用户在子商户appid下的唯一标识
	SubIsSubscribe bool   // 用户是否关注子公众账号，仅在公众账号类型支付有效

	TradeType string
	BankType  string // 银行类型，采用字符串类型的银行标识

	TotalFee           int    // 订单总金额，单位为分
	FeeType            string // 货币类型，符合ISO 4217标准的三位字母代码，默认人民币：CNY
	CashFee            int    // 订单现金支付金额，单位为分
	CashFeeType        string // 货币类型，符合ISO 4217标准的三位字母代码，默认人民币：CNY
	SettlementTotalFee int    // 应结订单金额=订单金额-非充值代金券金额，应结订单金额<=订单金额
	CouponFee          int    // 代金券或立减优惠金额<=订单总金额，订单总金额-代金券或立减优惠金额=现金支付金额
	CouponCount        int    // 代金券或立减优惠使用数量
	// TODO: coupon list

	TransactionID string    // 微信支付订单号
	OutTradeNo    string    // 商户系统内部订单号，要求32个字符内，只能是数字、大小写字母_-|*@ ，且在同一个商户号下唯一
	Attach        string    // 附加数据，在查询API和支付通知中原样返回，该字段主要用于商户携带订单的自定义数据
	TimeEnd       time.Time // 支付完成时间，格式为yyyyMMddHHmmss，如2009年12月25日9点10分10秒表示为20091225091010
}

func getOrderInfo(req map[string]string) (*OrderInfo, error) {
	totalFee, err := strconv.Atoi(req["total_fee"])
	if err != nil {
		return nil, err
	}
	cashFee, err := strconv.Atoi(req["cash_fee"])
	if err != nil {
		return nil, err
	}
	var settlementTotalFee, couponFee, couponCount int
	if req["settlement_total_fee"] != "" {
		settlementTotalFee, err = strconv.Atoi(req["settlement_total_fee"])
		if err != nil {
			return nil, err
		}
	}
	if req["coupon_fee"] != "" {
		couponFee, err = strconv.Atoi(req["coupon_fee"])
		if err != nil {
			return nil, err
		}
	}
	if req["coupon_count"] != "" {
		couponCount, err = strconv.Atoi(req["coupon_count"])
		if err != nil {
			return nil, err
		}
	}
	timeEnd, err := ParseTime(req["time_end"])
	if err != nil {
		return nil, err
	}

	return &OrderInfo{
		DeviceInfo:         req["device_info"],
		OpenID:             req["openid"],
		IsSubscribe:        req["is_subscribe"] == "Y",
		SubOpenID:          req["sub_openid"],
		SubIsSubscribe:     req["sub_is_subscribe"] == "Y",
		TradeType:          req["trade_type"],
		BankType:           req["bank_type"],
		TotalFee:           totalFee,
		FeeType:            req["fee_type"],
		CashFee:            cashFee,
		CashFeeType:        req["cash_fee_type"],
		SettlementTotalFee: settlementTotalFee,
		CouponFee:          couponFee,
		CouponCount:        couponCount,
		TransactionID:      req["transaction_id"],
		OutTradeNo:         req["out_trade_no"],
		Attach:             req["attach"],
		TimeEnd:            timeEnd,
	}, nil
}

type QueryOrderResponse struct {
	TradeState            // 交易状态
	TradeStateDesc string // 对当前查询订单状态的描述和下一步操作的指引

	OrderInfo

	Detail string // TODO: 商品详细列表
}

func (client *Client) QueryOrder(req *QueryOrderRequest) (rep *QueryOrderResponse, err error) {
	reqMap := make(map[string]string)
	if req.TransactionID != "" {
		reqMap["transaction_id"] = req.TransactionID
	}
	if req.OutTradeNo != "" {
		reqMap["out_trade_no"] = req.OutTradeNo
	}

	repMap, err := client.PostXML("/pay/orderquery", reqMap)
	if err != nil {
		client.Errorf("QueryOrder error: %s", err)
		return nil, err
	}
	client.Infof("QueryOrder response map: %s", repMap)

	tradeState := TradeState(repMap["trade_state"])
	if tradeState != TradeStateSUCCESS {
		rep = &QueryOrderResponse{
			TradeState:     tradeState,
			TradeStateDesc: repMap["trade_state_desc"],
			OrderInfo: OrderInfo{
				OutTradeNo: repMap["out_trade_no"],
				Attach:     repMap["attach"],
			},
		}
		return rep, nil
	}

	orderInfo, err := getOrderInfo(repMap)
	if err != nil {
		return nil, err
	}

	rep = &QueryOrderResponse{
		TradeState:     tradeState,
		TradeStateDesc: repMap["trade_state_desc"],
		OrderInfo:      *orderInfo,
		Detail:         repMap["detail"],
	}

	if req.TransactionID != "" && rep.TransactionID != "" && req.TransactionID != rep.TransactionID {
		err = fmt.Errorf("transaction_id mismatch, have: %s, want: %s", rep.TransactionID, req.TransactionID)
		return nil, err
	}
	if req.OutTradeNo != "" && rep.OutTradeNo != "" && req.OutTradeNo != rep.OutTradeNo {
		err = fmt.Errorf("out_trade_no mismatch, have: %s, want: %s", rep.OutTradeNo, req.OutTradeNo)
		return nil, err
	}

	return rep, nil
}
