package mch

import (
	"fmt"
)

const (
	ReturnCodeSuccess = "SUCCESS"
	ReturnCodeFail    = "FAIL"
)

const (
	ResultCodeSuccess = "SUCCESS"
	ResultCodeFail    = "FAIL"
)

type Error struct {
	ReturnCode string
	ReturnMsg  string
}

func (e *Error) Error() string {
	return fmt.Sprintf("return_code: %q, return_msg: %q", e.ReturnCode, e.ReturnMsg)
}

type BizError struct {
	ResultCode  string
	ErrCode     string
	ErrCodeDesc string
}

func (e *BizError) Error() string {
	return fmt.Sprintf("result_code: %q, err_code: %q, err_code_des: %q", e.ResultCode, e.ErrCode, e.ErrCodeDesc)
}

const (
	ErrCodeSYSTEMERROR = "SYSTEMERROR" // 系统异常，请用相同参数重新调用

	ErrCodeORDERPAID     = "ORDERPAID"     // 订单已支付，请当作已支付的正常交易
	ErrCodeORDERCLOSED   = "ORDERCLOSED"   // 订单已关闭
	ErrCodeNOTENOUGH     = "NOTENOUGH"     // 用户帐号余额不足，请用户充值或更换支付卡后再支付
	ErrCodeORDERNOTEXIST = "ORDERNOTEXIST" // 查询系统中不存在此交易订单号

	ErrCodeBIZERR_NEED_RETRY     = "BIZERR_NEED_RETRY"     // 并发情况下业务被拒绝，商户重试即可解决
	ErrCodeERROR                 = "ERROR"                 // 申请退款业务发生错误
	ErrCodeUSER_ACCOUNT_ABNORMAL = "USER_ACCOUNT_ABNORMAL" // 退款请求失败，用户帐号注销
	ErrCodeINVALID_REQ_TOO_MUCH  = "INVALID_REQ_TOO_MUCH"  // 连续错误请求数过多被系统短暂屏蔽，请在1分钟后再来重试
	ErrCodeFREQUENCY_LIMITED     = "FREQUENCY_LIMITED"     // 2个月之前的订单申请退款有频率限制，请降低频率后重试
	ErrCodeREFUNDNOTEXIST        = "REFUNDNOTEXIST"        // 请检查订单号是否有误以及订单状态是否正确，如：未支付、已支付未退款

	ErrCodeNOAUTH                = "NOAUTH" // 商户未开通此接口权限	请商户前往申请此接口权限
	ErrCodeAPPID_NOT_EXIST       = "APPID_NOT_EXIST"
	ErrCodeMCHID_NOT_EXIST       = "MCHID_NOT_EXIST"
	ErrCodeAPPID_MCHID_NOT_MATCH = "APPID_MCHID_NOT_MATCH"
	ErrCodeLACK_PARAMS           = "LACK_PARAMS"
	ErrCodeOUT_TRADE_NO_USED     = "OUT_TRADE_NO_USED" // 商户订单号重复，同一笔交易不能多次提交，请核实商户订单号是否重复提交
	ErrCodeSIGNERROR             = "SIGNERROR"
	ErrCodeXML_FORMAT_ERROR      = "XML_FORMAT_ERROR"
	ErrCodeREQUIRE_POST_METHOD   = "REQUIRE_POST_METHOD"
	ErrCodePOST_DATA_EMPTY       = "POST_DATA_EMPTY"
	ErrCodeNOT_UTF8              = "NOT_UTF8"
	ErrCodeTRADE_OVERDUE         = "TRADE_OVERDUE"         // 订单已经超过可退款的最大期限(支付后一年内可退款)
	ErrCodeINVALID_TRANSACTIONID = "INVALID_TRANSACTIONID" // 检查原交易号是否存在或发起支付交易接口返回失败
	ErrCodePARAM_ERROR           = "PARAM_ERROR"

	ErrCodeNO_AUTH               = "NO_AUTH"               // 没有授权请求此api
	ErrCodeAMOUNT_LIMIT          = "AMOUNT_LIMIT"          // 付款金额不能小于最低限额	每次付款金额必须大于1元
	ErrCodeOPENID_ERROR          = "OPENID_ERROR"          // Openid格式错误或者不属于商家公众账号
	ErrCodeSEND_FAILED           = "SEND_FAILED"           // 付款失败，请换单号重试
	ErrCodeNAME_MISMATCH         = "NAME_MISMATCH"         // 请求参数里填写了需要检验姓名，但是输入了错误的姓名
	ErrCodeFREQ_LIMIT            = "FREQ_LIMIT"            // 接口请求频率超时接口限制
	ErrCodeMONEY_LIMIT           = "MONEY_LIMIT"           // 已经达到今日付款总额上限/已达到付款给此用户额度上限
	ErrCodeV2_ACCOUNT_SIMPLE_BAN = "V2_ACCOUNT_SIMPLE_BAN" // 无法给非实名用户付款
	ErrCodeNOT_FOUND             = "NOT_FOUND"
)
