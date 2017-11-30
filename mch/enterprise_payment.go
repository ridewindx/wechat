package mch

import "strconv"

type EnterprisePayRequst struct {
	// 必选参数
	OutTradeNo  string // 商户订单号，需保持唯一性(只能是字母或者数字，不能包含有符号)
	Amount      int    // 企业付款金额，单位为分
	Desc        string // 企业付款操作说明信息
	FromIP      string // 调用接口的机器Ip地址
	OpenID      string // 商户appid下，某用户的openid
	IsCheckName bool   // 是否校验真实姓名

	// 可选参数
	UserName   string // 收款用户真实姓名，如果check_name设置为FORCE_CHECK，则必填用户真实姓名
	DeviceInfo string // 微信支付分配的终端设备号
}

type EnterprisePayResponse struct {
	PaymentNo   string // 企业付款成功，返回的微信订单号
	PaymentTime string // 企业付款成功时间
	OutTradeNo  string // 商户订单号
	DeviceInfo  string // 调用接口提交的终端设备号
}

func (client *Client) EnterprisePay(req *EnterprisePayRequst) (*EnterprisePayResponse, error) {
	reqMap := map[string]string{
		"partner_trade_no": req.OutTradeNo,
		"amount":           strconv.Itoa(req.Amount),
		"desc":             req.Desc,
		"spbill_create_ip": req.FromIP,
		"openid":           req.OpenID,
	}

	if req.IsCheckName {
		reqMap["check_name"] = "FORCE_CHECK"
		reqMap["re_user_name"] = req.UserName
	} else {
		reqMap["check_name"] = "NO_CHECK"
	}
	if req.DeviceInfo != "" {
		reqMap["device_info"] = req.DeviceInfo
	}

	reqMap["mch_appid"] = client.appID
	reqMap["mchid"] = client.mchID

	repMap, err := client.PostXMLWithCert("/mmpaymkttransfers/promotion/transfers", reqMap)
	if err != nil {
		return nil, err
	}

	rep := &EnterprisePayResponse{
		PaymentNo:   repMap["payment_no"],
		PaymentTime: repMap["payment_time"],
		OutTradeNo:  repMap["partner_trade_no"],
		DeviceInfo:  repMap["device_info"],
	}
	return rep, nil
}

const (
	EnterprisePaymentSuccess    = "SUCCESS"
	EnterprisePaymentFailed     = "FAILED"
	EnterprisePaymentProcessing = "PROCESSING"
)

type QueryEnterprisePaymentResponse struct {
	Status      string // 转账状态
	Reason      string // 如果失败则有失败原因
	PaymentNo   string // 微信订单号
	Amount      int    // 付款金额单位分
	Desc        string // 付款时候的描述
	PaymentTime string // 发起转账的时间
	OpenID      string // 商户appid下用户的openid
	UserName    string // 收款用户真实姓名
}

func (client *Client) QueryEnterprisePayment(outTradeNo string) (*QueryEnterprisePaymentResponse, error) {
	reqMap := map[string]string{
		"partner_trade_no": outTradeNo,
	}
	repMap, err := client.PostXML("/mmpaymkttransfers/gettransferinfo", reqMap)
	if err != nil {
		return nil, err
	}
	amount, err := strconv.Atoi(repMap["payment_amount"])
	if err != nil {
		return nil, err
	}
	rep := &QueryEnterprisePaymentResponse{
		Status:      repMap["status"],
		Reason:      repMap["reason"],
		PaymentNo:   repMap["detail_id"],
		Amount:      amount,
		Desc:        repMap["desc"],
		PaymentTime: repMap["transfer_time"],
		OpenID:      repMap["openid"],
		UserName:    repMap["transfer_name"],
	}
	return rep, nil
}
