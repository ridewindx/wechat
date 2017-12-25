package mch

import (
	"net/http"
	"crypto/md5"
	"crypto/hmac"
	"fmt"
	"crypto/sha256"
	"go.uber.org/zap"
	"github.com/jiudaoyun/wechat"
	"io"
	"encoding/base64"
	"github.com/nanjishidu/gomini/gocrypto"
	"bytes"
	"strconv"
	"time"
)

type NotifyMsg struct {
	AppID    string
	MchID    string
	SubAppID string
	SubMchID string

	OrderInfo
}

type RefundNotifyMsg struct {
	TransactionID string
	OutTradeNO string
	RefundID string
	OutRefundNO string
	TotalFee int
	RefundFee int
	RefundStatus RefundStatus
	SuccessTime time.Time
}

type NotifyHandler struct {
	payNotifyHandler func(*NotifyMsg) error
	refundNotifyHandler func(*RefundNotifyMsg) error

	appID  string
	mchID  string
	apiKey string

	subAppID string
	subMchID string

	*zap.SugaredLogger
}

func NewNotifyHandler(appID, mchID, apiKey string, payNofityHandler func(*NotifyMsg) error, refundNotifyHandler func(*RefundNotifyMsg) error) *NotifyHandler {
	apiKeyMD5 := fmt.Sprintf("%x", md5.Sum([]byte(apiKey)))
	gocrypto.SetAesKey(apiKeyMD5)

	return &NotifyHandler{
		payNotifyHandler: payNofityHandler,
		refundNotifyHandler: refundNotifyHandler,

		appID:  appID,
		mchID:  mchID,
		apiKey: apiKey,

		SugaredLogger: wechat.Sugar,
	}
}

func (nm *NotifyHandler) reply(w io.Writer, code, reason string) {
	err := EncodeXML(w, map[string]string{
		"return_code": code,
		"return_msg":  reason,
	})
	if err != nil {
		nm.Errorw("NotifyHandler write XML failed", "error", err)
	}
}

func (nm *NotifyHandler) replyError(w io.Writer, reason string) {
	nm.Errorw("NotifyHandler reply error", "error", reason)
	nm.reply(w, ReturnCodeFail, reason)
}

func (nm *NotifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	nm.Serve(w, r, nm.payNotifyHandler, nm.refundNotifyHandler)
}

func (nm *NotifyHandler) Serve(w io.Writer, r *http.Request, payNotifyHandler func(*NotifyMsg) error, refundNotifyHandler func(*RefundNotifyMsg) error) {
	if r.Method != "POST" {
		nm.replyError(w, "unexpected HTTP Method: "+r.Method)
		return
	}

	req, err := DecodeXML(r.Body)
	if err != nil {
		nm.replyError(w, err.Error())
		return
	}

	nm.Infof("wechat callback: %s", req)

	returnCode := req["return_code"]
	if returnCode != ReturnCodeSuccess {
		err = &Error{
			ReturnCode: returnCode,
			ReturnMsg:  req["return_msg"],
		}
		nm.replyError(w, err.Error())
		return
	}

	if nm.appID != "" {
		wantAppId := nm.appID
		haveAppId := req["appid"]
		if !wechat.Compare(haveAppId, wantAppId) {
			nm.replyError(w, fmt.Sprintf("appid mismatch, have: %s, want: %s", haveAppId, wantAppId))
			return
		}
	}
	if nm.mchID != "" {
		wantMchId := nm.mchID
		haveMchId := req["mch_id"]
		if !wechat.Compare(haveMchId, wantMchId) {
			nm.replyError(w, fmt.Sprintf("mch_id mismatch, have: %s, want: %s", haveMchId, wantMchId))
			return
		}
	}

	if nm.subAppID != "" {
		wantSubAppId := nm.subAppID
		haveSubAppId := req["sub_appid"]
		if haveSubAppId != "" && !wechat.Compare(haveSubAppId, wantSubAppId) {
			nm.replyError(w, fmt.Sprintf("sub_appid mismatch, have: %s, want: %s", haveSubAppId, wantSubAppId))
			return
		}
	}
	if nm.subMchID != "" {
		wantSubMchId := nm.subMchID
		haveSubMchId := req["sub_mch_id"]
		if !wechat.Compare(haveSubMchId, wantSubMchId) {
			nm.replyError(w, fmt.Sprintf("sub_mch_id mismatch, have: %s, want: %s", haveSubMchId, wantSubMchId))
			return
		}
	}

	// 1. refund order notify
	if reqInfo, ok := req["req_info"]; ok {
		encrypted, err := base64.StdEncoding.DecodeString(reqInfo)
		if err != nil {
			nm.replyError(w, err.Error())
			return
		}
		decrypted, err := gocrypto.AesECBDecrypt(encrypted)
		if err != nil {
			nm.replyError(w, err.Error())
			return
		}
		// nm.Infof("wechat callback: %s", decrypted)
		info, err := DecodeXML(bytes.NewBuffer(decrypted))
		if err != nil {
			nm.replyError(w, err.Error())
			return
		}

		msg, err := getRefundNotifyMsg(info)
		if err != nil {
			nm.replyError(w, err.Error())
			return
		}

		err = refundNotifyHandler(msg)
		if err != nil {
			nm.replyError(w, err.Error())
			return
		}

		nm.reply(w, ReturnCodeSuccess, "OK")
		return
	}

	// 2. pay order notify

	haveSign := req["sign"]
	var wantSign string
	switch signType := req["sign_type"]; signType {
	case "", SignTypeMD5:
		wantSign = Sign(req, nm.apiKey, md5.New())
	case SignTypeHMAC_SHA256:
		wantSign = Sign(req, nm.apiKey, hmac.New(sha256.New, []byte(nm.apiKey)))
	default:
		nm.replyError(w, fmt.Sprintf("unsupported notification sign_type: %s", signType))
		return
	}
	if !wechat.Compare(haveSign, wantSign) {
		nm.replyError(w, fmt.Sprintf("sign mismatch,\nhave: %s,\nwant: %s", haveSign, wantSign))
		return
	}

	resultCode := req["result_code"]
	if resultCode != ResultCodeSuccess {
		err = &BizError{
			ResultCode:  resultCode,
			ErrCode:     req["err_code"],
			ErrCodeDesc: req["err_code_des"],
		}
		nm.replyError(w, err.Error())
		return
	}

	orderInfo, err := getOrderInfo(req)
	if err != nil {
		nm.replyError(w, err.Error())
		return
	}

	msg := NotifyMsg{
		AppID: req["appid"],
		MchID: req["mch_id"],
		SubAppID: req["sub_appid"],
		SubMchID: req["sub_mch_id"],
		OrderInfo: *orderInfo,
	}
	err = payNotifyHandler(&msg)
	if err != nil {
		nm.replyError(w, err.Error())
		return
	}

	nm.reply(w, ReturnCodeSuccess, "OK")
}

func getRefundNotifyMsg(req map[string]string) (*RefundNotifyMsg, error) {
	totalFee, err := strconv.Atoi(req["total_fee"])
	if err != nil {
		return nil, err
	}
	refundFee, err := strconv.Atoi(req["refund_fee"])
	if err != nil {
		return nil, err
	}
	var successTime time.Time
	if str, ok := req["success_time"]; ok {
		successTime, err = time.Parse("2006-01-02 15:04:05", str)
		if err != nil {
			return nil, err
		}
	}

	return &RefundNotifyMsg{
		TransactionID: req["transaction_id"],
		OutTradeNO: req["out_trade_no"],
		RefundID: req["refund_id"],
		OutRefundNO: req["out_refund_no"],
		TotalFee: totalFee,
		RefundFee: refundFee,
		RefundStatus: RefundStatus(req["refund_status"]),
		SuccessTime: successTime,
	}, nil
}
