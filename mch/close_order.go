package mch

type CloseOrderRequest struct {
	OutTradeNo string `xml:"out_trade_no"` // 商户系统内部订单号
}

// 以下情况需要调用关单接口：
// 商户订单支付失败需要生成新单号重新发起支付，要对原订单号调用关单，避免重复支付；
// 系统下单后，用户支付超时，系统退出不再受理，避免用户继续，请调用关单接口。
func CloseOrder(clt *Client, req *CloseOrderRequest) (err error) {
	reqMap := make(map[string]string)
	reqMap["out_trade_no"] = req.OutTradeNo

	_, err = clt.PostXML("/pay/closeorder", reqMap)
	return
}
