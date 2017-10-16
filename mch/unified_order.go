package mch

import (
	"time"
	"strconv"
	"fmt"
)

type UnifiedOrderRequest struct {
	// 必选参数
	Body           string // 商品或支付单简要描述
	OutTradeNo     string // 商户系统内部的订单号,32个字符内、可包含字母, 其他说明见商户订单号
	TotalFee       int  // 订单总金额，单位为分，详见支付金额
	SpbillCreateIP string // APP和网页支付提交用户端ip，Native支付填调用微信支付API的机器IP。
	NotifyURL      string // 接收微信支付异步通知回调地址，通知url必须为直接可访问的url，不能携带参数。
	TradeType      string // 取值如下：JSAPI，NATIVE，APP，详细说明见参数规定

	// 可选参数
	DeviceInfo string // 终端设备号(门店号或收银设备ID)，注意：PC网页或公众号内支付请传"WEB"
	Detail     string // 商品名称明细列表
	Attach     string // 附加数据，在查询API和支付通知中原样返回，该字段主要用于商户携带订单的自定义数据
	FeeType    string // 符合ISO 4217标准的三位字母代码，默认人民币：CNY，其他值列表详见货币类型
	TimeStart  time.Time // 订单生成时间，格式为yyyyMMddHHmmss，如2009年12月25日9点10分10秒表示为20091225091010。其他详见时间规则
	TimeExpire time.Time // 订单失效时间，格式为yyyyMMddHHmmss，如2009年12月27日9点10分10秒表示为20091227091010。其他详见时间规则
	GoodsTag   string // 商品标记，代金券或立减优惠功能的参数，说明详见代金券或立减优惠
	ProductId  string // trade_type=NATIVE，此参数必传。此id为二维码中包含的商品ID，商户自行定义。
	LimitPay   string // no_credit--指定不能使用信用卡支付
	OpenId     string // rade_type=JSAPI，此参数必传，用户在商户appid下的唯一标识。
	SubOpenId  string // trade_type=JSAPI，此参数必传，用户在子商户appid下的唯一标识。openid和sub_openid可以选传其中之一，如果选择传sub_openid,则必须传sub_appid。
	SceneInfo  string // 该字段用于上报支付的场景信息,针对H5支付有以下三种场景,请根据对应场景上报,H5支付不建议在APP端使用，针对场景1，2请接入APP支付，不然可能会出现兼容性问题
}

type UnifiedOrderResponse struct {
	PrepayId  string // 微信生成的预支付回话标识，用于后续接口调用中使用，该值有效期为2小时
	TradeType string // 调用接口提交的交易类型，取值如下：JSAPI，NATIVE，APP，详细说明见参数规定

	// 下面字段都是可选返回的(详细见微信支付文档), 为空值表示没有返回, 程序逻辑里需要判断
	DeviceInfo string // 调用接口提交的终端设备号
	CodeURL    string // trade_type 为 NATIVE 时有返回，可将该参数值生成二维码展示出来进行扫码支付
	MWebURL    string // trade_type 为 MWEB 时有返回
}

func (client *Client) UnifiedOrder(req *UnifiedOrderRequest) (rep *UnifiedOrderResponse, err error) {
	reqMap := make(map[string]string, 24)
	reqMap["body"] = req.Body
	reqMap["out_trade_no"] = req.OutTradeNo
	reqMap["total_fee"] = strconv.Itoa(req.TotalFee)
	reqMap["spbill_create_ip"] = req.SpbillCreateIP
	reqMap["notify_url"] = req.NotifyURL
	reqMap["trade_type"] = req.TradeType
	if req.DeviceInfo != "" {
		reqMap["device_info"] = req.DeviceInfo
	}
	if req.Detail != "" {
		reqMap["detail"] = req.Detail
	}
	if req.Attach != "" {
		reqMap["attach"] = req.Attach
	}
	if req.FeeType != "" {
		reqMap["fee_type"] = req.FeeType
	}
	if !req.TimeStart.IsZero() {
		reqMap["time_start"] = FormatTime(req.TimeStart)
	}
	if !req.TimeExpire.IsZero() {
		reqMap["time_expire"] = FormatTime(req.TimeExpire)
	}
	if req.GoodsTag != "" {
		reqMap["goods_tag"] = req.GoodsTag
	}
	if req.ProductId != "" {
		reqMap["product_id"] = req.ProductId
	}
	if req.LimitPay != "" {
		reqMap["limit_pay"] = req.LimitPay
	}
	if req.OpenId != "" {
		reqMap["openid"] = req.OpenId
	}
	if req.SubOpenId != "" {
		reqMap["sub_openid"] = req.SubOpenId
	}
	if req.SceneInfo != "" {
		reqMap["scene_info"] = req.SceneInfo
	}

	repMap, err := client.PostXML("/pay/unifiedorder", reqMap)
	if err != nil {
		return nil, err
	}

	repTradeType := repMap["trade_type"]
	if repTradeType != req.TradeType {
		err = fmt.Errorf("trade_type mismatch, have: %s, want: %s", repTradeType, req.TradeType)
		return nil, err
	}

	rep = &UnifiedOrderResponse{
		PrepayId:   repMap["prepay_id"],
		TradeType:  repTradeType,
		DeviceInfo: repMap["device_info"],
		CodeURL:    repMap["code_url"],
		MWebURL:    repMap["mweb_url"],
	}
	return rep, nil
}
