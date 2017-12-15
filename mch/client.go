package mch

import (
	"net/http"
	"time"
	"crypto/md5"
	"crypto/hmac"
	"fmt"
	"crypto/sha256"
	"bytes"
	"sync"
	"strings"
	"io"
	"crypto/tls"
	"go.uber.org/zap"
	"github.com/jiudaoyun/wechat"
)

const API_URL = "https://api.mch.weixin.qq.com"

type Client struct {
	appID  string
	mchID  string
	apiKey string

	subAppID string
	subMchID string

	signType string // 签名类型，目前支持HMAC-SHA256和MD5，默认为MD5

	client    *http.Client
	tlsClient *http.Client

	*zap.SugaredLogger
}

func New(appID, mchID, apiKey string, timeout ...time.Duration) *Client {
	client := &Client{
		appID:  appID,
		mchID:  mchID,
		apiKey: apiKey,

		SugaredLogger: wechat.Sugar,
	}
	if len(timeout) > 0 {
		client.client = &http.Client{
			Timeout: timeout[0],
		}
	}
	return client
}

func NewWithSubMch(appID, mchID, apiKey, subAppID, subMchID string, timeout ...time.Duration) *Client {
	client := &Client{
		appID:    appID,
		mchID:    mchID,
		apiKey:   apiKey,
		subAppID: subAppID,
		subMchID: subMchID,

		SugaredLogger: wechat.Sugar,
	}
	if len(timeout) > 0 {
		client.client = &http.Client{
			Timeout: timeout[0],
		}
	}
	return client
}

func (client *Client) SetCert(certPEMBlock, keyPEMBlock []byte) error {
	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return err
	}
	return client.setCert(cert)
}

func (client *Client) SetCertFromFile(certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	return client.setCert(cert)
}

func (client *Client) setCert(cert tls.Certificate) error {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	transport := *http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = tlsConfig
	client.tlsClient = &http.Client{
		Transport: &transport,
		Timeout:   client.client.Timeout,
	}
	return nil
}

func (client *Client) SetSignType(signType string) error {
	if signType != SignTypeHMAC_SHA256 && signType != SignTypeMD5 {
		return fmt.Errorf("unsupported sign type: %s", signType)
	}
	client.signType = signType
	return nil
}

var pool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 4<<10)) // 4KB
	},
}

func (client *Client) PostXML(relativeURL string, req map[string]string) (rep map[string]string, err error) {
	return client.postXML(false, relativeURL, req)
}

func (client *Client) PostXMLWithCert(relativeURL string, req map[string]string) (rep map[string]string, err error) {
	return client.postXML(true, relativeURL, req)
}

func (client *Client) postXML(withCert bool, relativeURL string, req map[string]string) (rep map[string]string, err error) {
	_, isEnterprisePay := req["mch_appid"]

	buffer, err := client.makeRequest(req, isEnterprisePay)
	defer pool.Put(buffer)
	if err != nil {
		return nil, err
	}

	url := API_URL + relativeURL
	rep, err = client.postToMap(withCert, url, buffer, isEnterprisePay)
	if err != nil {
		bizError, ok := err.(*BizError)
		if !ok || (bizError.ErrCode != ErrCodeSYSTEMERROR && bizError.ErrCode != ErrCodeBIZERR_NEED_RETRY) {
			return
		}
		url = switchReqURL(url)
		return client.postToMap(withCert, url, buffer, isEnterprisePay) // retry
	}
	return
}

func (client *Client) makeRequest(req map[string]string, isEnterprisePay bool) (*bytes.Buffer, error) {
	if !isEnterprisePay {
		req["appid"] = client.appID
		req["mch_id"] = client.mchID
	}
	if client.subAppID != "" {
		req["sub_appid"] = client.subAppID
	}
	if client.subMchID != "" {
		req["sub_mch_id"] = client.subMchID
	}

	req["nonce_str"] = RandString(32)

	switch client.signType {
	case "", SignTypeMD5:
		if !isEnterprisePay {
			req["sign_type"] = SignTypeMD5
		}
		req["sign"] = Sign(req, client.apiKey, md5.New())
	case SignTypeHMAC_SHA256:
		if !isEnterprisePay {
			req["sign_type"] = SignTypeHMAC_SHA256
		}
		req["sign"] = Sign(req, client.apiKey, hmac.New(sha256.New, []byte(client.apiKey)))
	}
	fmt.Printf("req: %v\n", req)

	buffer := pool.Get().(*bytes.Buffer)
	buffer.Reset()

	err := EncodeXML(buffer, req)
	return buffer, err
}

func (client *Client) postToMap(withCert bool, url string, body io.Reader, isEnterprisePay bool) (rep map[string]string, err error) {
	repBody, err := client.post(withCert, url, body)
	if err != nil {
		return nil, err
	}
	defer repBody.Close()
	return client.toMap(repBody, isEnterprisePay)
}

func (client *Client) post(withCert bool, url string, body io.Reader) (io.ReadCloser, error) {
	c := client.client
	if withCert {
		c = client.tlsClient
	}
	rep, err := c.Post(url, "text/xml; charset=utf-8", body)
	if err != nil {
		return nil, err
	}
	if rep.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status: %s", rep.Status)
	}

	return rep.Body, nil
}

func (client *Client) toMap(repBody io.Reader, isEnterprisePay bool) (rep map[string]string, err error) {
	if closer, ok := repBody.(io.Closer); ok {
		defer closer.Close()
	}

	rep, err = DecodeXML(repBody)
	if err != nil {
		return nil, err
	}
	fmt.Printf("toMap rep: %v\n", rep)

	returnCode := rep["return_code"]
	if returnCode != ReturnCodeSuccess {
		return nil, &Error{
			ReturnCode: returnCode,
			ReturnMsg:  rep["return_msg"],
		}
	}

	resultCode := rep["result_code"]
	if resultCode != ResultCodeSuccess {
		return nil, &BizError{
			ResultCode:  resultCode,
			ErrCode:     rep["err_code"],
			ErrCodeDesc: rep["err_code_des"],
		}
	}

	appId := rep["appid"]
	if appId != "" && appId != client.appID {
		return nil, fmt.Errorf("appid mismatch, have: %s, want: %s", appId, client.appID)
	}
	mchId := rep["mch_id"]
	if mchId == "" {
		mchId, _ = rep["mchid"]
	}
	if mchId != "" && mchId != client.mchID {
		return nil, fmt.Errorf("mch_id mismatch, have: %s, want: %s", mchId, client.mchID)
	}

	if client.subAppID != "" {
		subAppId := rep["sub_appid"]
		subMchId := rep["sub_mch_id"]
		if subAppId != "" && subAppId != client.subAppID {
			return nil, fmt.Errorf("sub_appid mismatch, have: %s, want: %s", subAppId, client.subAppID)
		}
		if subMchId != client.subMchID {
			return nil, fmt.Errorf("sub_mch_id mismatch, have: %s, want: %s", subMchId, client.subMchID)
		}
	}

	if isEnterprisePay { // enterprise payment response has no sign
		return rep, nil
	}

	var signWant string
	signHave := rep["sign"]
	repSignType := rep["sign_type"]
	switch repSignType {
	case "", SignTypeMD5:
		signWant = Sign(rep, client.apiKey, md5.New())
	case SignTypeHMAC_SHA256:
		signWant = Sign(rep, client.apiKey, hmac.New(sha256.New, []byte(client.apiKey)))
	default:
		err = fmt.Errorf("unsupported response sign_type: %s", repSignType)
		return nil, err
	}
	if signHave != signWant {
		return nil, fmt.Errorf("sign mismatch,\nhave: %s,\nwant: %s", signHave, signWant)
	}

	return rep, nil
}

func switchReqURL(url string) string {
	switch {
	case strings.HasPrefix(url, "https://api.mch.weixin.qq.com/"):
		return "https://api2.mch.weixin.qq.com/" + url[len("https://api.mch.weixin.qq.com/"):]
	case strings.HasPrefix(url, "https://api2.mch.weixin.qq.com/"):
		return "https://api.mch.weixin.qq.com/" + url[len("https://api2.mch.weixin.qq.com/"):]
	default:
		return url
	}
}
