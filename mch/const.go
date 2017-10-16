package mch

type TradeType string

const (
	TradeTypeMWEB TradeType = "MWEB" // H5支付
	TradeTypeJSAPI TradeType = "JSAPI" // 公众号支付/小程序支付
	TradeTypeAPP TradeType = "APP" // APP支付
	TradeTypeNATIVE TradeType = "NATIVE" // 原生扫码支付
	TradeTypeMICROPAY TradeType = "MICROPAY" // 刷卡支付
)

type TradeState string

const (
	TradeStateSUCCESS TradeState = "SUCCESS" // 支付成功
	TradeStateREFUND TradeState = "REFUND" // 转入退款
	TradeStateNOTPAY TradeState = "NOTPAY" // 未支付
	TradeStateCLOSED TradeState = "CLOSED" // 已关闭
	TradeStateREVOKED TradeState = "REVOKED" // 已撤销(刷卡支付)
	TradeStateUSERPAYING TradeState = "USERPAYING" // 用户支付中
	TradeStatePAYERROR TradeState = "PAYERROR" // 支付失败(其他原因，如银行返回失败)
)
