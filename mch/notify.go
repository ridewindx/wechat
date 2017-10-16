package mch

import (
	"net/http"
	"crypto/md5"
	"crypto/hmac"
	"fmt"
	"crypto/sha256"
	"go.uber.org/zap"
	"github.com/jiudaoyun/wechat"
)

type NotifyMsg struct {
	AppID    string
	MchID    string
	SubAppID string
	SubMchID string

	OrderInfo
}

type NotifyHandler struct {
	handler func(*NotifyMsg) error

	appID  string
	mchID  string
	apiKey string

	subAppID string
	subMchID string

	*zap.SugaredLogger
}

func NewNotifyHandler(appID, mchID, apiKey string, handler func(*NotifyMsg) error) *NotifyHandler {
	return &NotifyHandler{
		handler: handler,

		appID:  appID,
		mchID:  mchID,
		apiKey: apiKey,

		SugaredLogger: wechat.Sugar,
	}
}

func (nm *NotifyHandler) serveError(w http.ResponseWriter, reason string) {
	nm.Errorw("NotifyHandler serve error", "error", reason)
	err := EncodeXML(w, map[string]string{
		"return_code": ReturnCodeFail,
		"return_msg":  reason,
	})
	if err != nil {
		nm.Errorw("NotifyHandler write XML failed", "error", err)
	}
}

func (nm *NotifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		nm.serveError(w, "unexpected HTTP Method: "+r.Method)
		return
	}

	req, err := DecodeXML(r.Body)
	if err != nil {
		nm.serveError(w, err.Error())
		return
	}

	returnCode := req["return_code"]
	if returnCode != ReturnCodeSuccess {
		err = &Error{
			ReturnCode: returnCode,
			ReturnMsg:  req["return_msg"],
		}
		nm.serveError(w, err.Error())
		return
	}

	resultCode := req["result_code"]
	if resultCode != ResultCodeSuccess {
		err = &BizError{
			ResultCode:  resultCode,
			ErrCode:     req["err_code"],
			ErrCodeDesc: req["err_code_des"],
		}
		nm.serveError(w, err.Error())
		return
	}

	if nm.appID != "" {
		wantAppId := nm.appID
		haveAppId := req["appid"]
		if !wechat.Compare(haveAppId, wantAppId) {
			nm.serveError(w, fmt.Sprintf("appid mismatch, have: %s, want: %s", haveAppId, wantAppId))
			return
		}
	}
	if nm.mchID != "" {
		wantMchId := nm.mchID
		haveMchId := req["mch_id"]
		if !wechat.Compare(haveMchId, wantMchId) {
			nm.serveError(w, fmt.Sprintf("mch_id mismatch, have: %s, want: %s", haveMchId, wantMchId))
			return
		}
	}

	if nm.subAppID != "" {
		wantSubAppId := nm.subAppID
		haveSubAppId := req["sub_appid"]
		if haveSubAppId != "" && !wechat.Compare(haveSubAppId, wantSubAppId) {
			nm.serveError(w, fmt.Sprintf("sub_appid mismatch, have: %s, want: %s", haveSubAppId, wantSubAppId))
			return
		}
	}
	if nm.subMchID != "" {
		wantSubMchId := nm.subMchID
		haveSubMchId := req["sub_mch_id"]
		if !wechat.Compare(haveSubMchId, wantSubMchId) {
			nm.serveError(w, fmt.Sprintf("sub_mch_id mismatch, have: %s, want: %s", haveSubMchId, wantSubMchId))
			return
		}
	}

	haveSign := req["sign"]
	var wantSign string
	switch signType := req["sign_type"]; signType {
	case "", SignTypeMD5:
		wantSign = Sign(req, nm.apiKey, md5.New())
	case SignTypeHMAC_SHA256:
		wantSign = Sign(req, nm.apiKey, hmac.New(sha256.New, []byte(nm.apiKey)))
	default:
		nm.serveError(w, fmt.Sprintf("unsupported notification sign_type: %s", signType))
		return
	}
	if !wechat.Compare(haveSign, wantSign) {
		nm.serveError(w, fmt.Sprintf("sign mismatch,\nhave: %s,\nwant: %s", haveSign, wantSign))
		return
	}

	orderInfo, err := getOrderInfo(req)
	if err != nil {
		nm.serveError(w, err.Error())
	}

	msg := NotifyMsg{
		AppID: req["appid"],
		MchID: req["mch_id"],
		SubAppID: req["sub_appid"],
		SubMchID: req["sub_mch_id"],
		OrderInfo: *orderInfo,
	}
	err = nm.handler(&msg)
	if err != nil {
		nm.serveError(w, err.Error())
	}

	err = EncodeXML(w, map[string]string{
		"return_code": ReturnCodeSuccess,
	})
	if err != nil {
		nm.Errorw("NotifyHandler write XML failed", "error", err)
	}
}
