package mp

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"github.com/ridewindx/mel"
	"github.com/ridewindx/mel/binding"
	"sort"
	"sync"
	"sync/atomic"
	"unsafe"
	"go.uber.org/zap"
	"github.com/jiudaoyun/wechat"
	"net/http"
	"fmt"
	"encoding/json"
	"strings"
	"github.com/ridewindx/melware"
)

type Server struct {
	*mel.Mel
	urlPrefix string

	appID string // App ID
	ID    string // Wechat ID

	tokenMutex sync.Mutex
	token      unsafe.Pointer

	aesKeyMutex sync.Mutex
	aesKey      unsafe.Pointer

	client            *Client
	middlewares       []Handler
	messageHandlerMap map[string]Handler
	eventHandlerMap   map[string]Handler

	logger *zap.SugaredLogger
}

func (srv *Server) setURLPrefix(urlPrefix string) {
	if !strings.HasPrefix(urlPrefix, "/") {
		urlPrefix = "/" + urlPrefix
	}
	urlPrefix = strings.TrimRight(urlPrefix, "/")
	srv.urlPrefix = urlPrefix
}

func (srv *Server) SetID(id string) {
	srv.ID = id
}

func (srv *Server) SetAppID(appID string) {
	srv.appID = appID
}

func (srv *Server) SetClient(client *Client) {
	srv.client = client
}

type Token struct {
	current string
	last    string
}

type AESKey struct {
	current string
	last    string
}

func (srv *Server) GetToken() (string, string) {
	p := (*Token)(atomic.LoadPointer(&srv.token))
	if p != nil {
		return p.current, p.last
	}
	return "", ""
}

func (srv *Server) SetToken(token string) {
	if token == "" {
		return
	}

	srv.tokenMutex.Lock()
	defer srv.tokenMutex.Unlock()

	current, _ := srv.GetToken()
	if token == current {
		return
	}

	t := Token{
		current: token,
		last:    current,
	}
	atomic.StorePointer(&srv.token, unsafe.Pointer(&t))
}

func (srv *Server) deleteLastToken() {
	srv.tokenMutex.Lock()
	defer srv.tokenMutex.Unlock()

	current, last := srv.GetToken()
	if last == "" {
		return
	}

	t := Token{
		current: current,
	}
	atomic.StorePointer(&srv.token, unsafe.Pointer(&t))
}

func (srv *Server) GetAESKey() (string, string) {
	p := (*AESKey)(atomic.LoadPointer(&srv.aesKey))
	if p != nil {
		return p.current, p.last
	}
	return "", ""
}

func (srv *Server) SetAESKey(base64AESKey string) {
	if len(base64AESKey) != 43 {
		return
	}
	aesKey, err := base64.StdEncoding.DecodeString(base64AESKey + "=")
	if err != nil {
		return
	}

	srv.aesKeyMutex.Lock()
	defer srv.aesKeyMutex.Unlock()

	current, _ := srv.GetAESKey()
	if bytes.Equal(aesKey, []byte(current)) {
		return
	}

	k := AESKey{
		current: string(aesKey),
		last:    current,
	}
	atomic.StorePointer(&srv.aesKey, unsafe.Pointer(&k))
}

func (srv *Server) deleteLastAESKey() {
	srv.aesKeyMutex.Lock()
	defer srv.aesKeyMutex.Unlock()

	current, last := srv.GetAESKey()
	if last == "" {
		return
	}

	k := AESKey{
		current: current,
	}
	atomic.StorePointer(&srv.aesKey, unsafe.Pointer(&k))
}

func (srv *Server) Use(middlewares ...Handler) {
	srv.middlewares = append(srv.middlewares, middlewares...)
	if len(srv.middlewares)+1 > int(abortIndex) {
		panic("too many middlewares")
	}
}

func (srv *Server) HandleMessage(msgType string, handler Handler) {
	srv.messageHandlerMap[msgType] = handler
}

func (srv *Server) HandleEvent(eventType string, handler Handler) {
	srv.eventHandlerMap[eventType] = handler
}

func (srv *Server) GetVerifyFile(filename string, content []byte) {
	srv.Get(srv.urlPrefix+"/"+filename, func(c *mel.Context) {
		c.Data(200, "text/plain", content)
	})
}

func NewServer(token, aesKey string, urlPrefix ...string) *Server {
	srv := &Server{
		Mel:               mel.New(),
		messageHandlerMap: make(map[string]Handler),
		eventHandlerMap:   make(map[string]Handler),
		logger:            wechat.Sugar,
	}

	srv.SetToken(token)
	srv.SetAESKey(aesKey)

	srv.Mel.Use(melware.Zap(srv.logger))

	cors := melware.CorsAllowAll()
	cors.AllowCredentials = false
	srv.Mel.Use(cors.Middleware())

	if len(urlPrefix) > 0 {
		srv.setURLPrefix(urlPrefix[0])
	}

	equal := func(a, b string) bool {
		return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
	}

	verifySignReturnToken := func(signature, timestamp, nonce string) string {
		currentToken, lastToken := srv.GetToken()
		token := currentToken

		isValid := func() bool {
			computedSignature := computeSign(token, timestamp, nonce)
			return equal(signature, computedSignature)
		}

		if isValid() {
			srv.deleteLastToken()
			return token
		}

		if lastToken != "" {
			token = lastToken
			if isValid() {
				return token
			}
		}

		return ""
	}

	verifySign := func(c *mel.Context) bool {
		signature := c.Query("signature")
		timestamp := c.Query("timestamp")
		nonce := c.Query("nonce")

		return verifySignReturnToken(signature, timestamp, nonce) != ""
	}

	type EncryptMsg struct {
		ToUserName string `xml:"ToUserName"`
		Encrypt    string `xml:"Encrypt"`
	}

	srv.Head("/", func(c *mel.Context) { // health check
		c.Status(200)
	})

	srv.Get(srv.urlPrefix+"/", func(c *mel.Context) {
		if verifySign(c) {
			echostr := c.Query("echostr")
			c.Text(200, echostr)
		}
	})

	handleMessage := func(event *Event) interface{} {
		var handler Handler
		var ok bool
		if event.Type == MessageEvent {
			handler, ok = srv.eventHandlerMap[event.Event]
		} else {
			handler, ok = srv.messageHandlerMap[event.Type]
		}
		if !ok {
			return nil // no registered handler, just respond with empty string
		}

		ctx := &Context{
			Client:   srv.client,
			index:    preStartIndex,
			handlers: append(srv.middlewares, handler),
			Event:    event,
		}

		ctx.Next()

		return ctx.response
	}

	srv.Post(srv.urlPrefix+"/", func(c *mel.Context) {
		encryptType := c.Query("encrypt_type")
		signature := c.Query("signature")
		timestamp := c.Query("timestamp")
		nonce := c.Query("nonce")

		switch encryptType {
		case "aes":
			token := verifySignReturnToken(signature, timestamp, nonce)
			if token == "" {
				srv.logger.Error("Verify sign empty token")
				return
			}

			msgSign := c.Query("msg_signature")

			var obj EncryptMsg
			err := c.BindWith(&obj, binding.XML)
			if err != nil {
				srv.logger.Errorw("Bind with XML failed", "error", err)
				return
			}

			if srv.ID != "" && !equal(obj.ToUserName, srv.ID) {
				srv.logger.Errorw("Wechat ID inconsistent", "id", srv.ID, "ToUserName", obj.ToUserName)
				return
			}

			computedSign := computeSign(token, timestamp, nonce, obj.Encrypt)
			if !equal(computedSign, msgSign) {
				srv.logger.Errorw("Signature inconsistent")
				return
			}

			encryptedMsg, err := base64.StdEncoding.DecodeString(obj.Encrypt)
			if err != nil {
				srv.logger.Errorw("Decode base64 string failed", "error", err)
				return
			}

			current, last := srv.GetAESKey()
			aesKey := current
			random, msg, appId, err := decryptMsg(encryptedMsg, []byte(aesKey))
			if err != nil {
				if last == "" {
					srv.logger.Errorw("Decrypt AES msg failed", "error", err)
					return
				}
				aesKey = last
				random, msg, appId, err = decryptMsg(encryptedMsg, []byte(aesKey))
				if err != nil {
					srv.logger.Errorw("Decrypt AES msg failed", "error", err)
					return
				}
			} else {
				srv.deleteLastAESKey()
			}
			if srv.appID != "" && string(appId) != srv.appID {
				srv.logger.Errorw("AppID inconsistent", "AppID", appId)
				return
			}

			var event Event
			if err = xml.Unmarshal(msg, &event); err != nil {
				srv.logger.Errorw("Unmarshal msg failed", "error", err)
				return
			}

			repBytes, err := xml.Marshal(handleMessage(&event))
			if err != nil {
				srv.logger.Errorw("Marshal msg failed", "error", err)
				return
			}

			encryptedRepBytes := encryptMsg(random, repBytes, appId, []byte(aesKey))
			encryptedRepStr := base64.StdEncoding.EncodeToString(encryptedRepBytes)
			repSignature := computeSign(token, timestamp, nonce, encryptedRepStr)

			type EncryptRepMsg struct {
				Encrypt      string
				MsgSignature string
				TimeStamp    string
				Nonce        string
			}

			err = c.XML(200, &EncryptRepMsg{encryptedRepStr, repSignature, timestamp, nonce})
			if err != nil {
				srv.logger.Errorw("Reply msg failed", "error", err)
			}

		case "", "raw":
			if !verifySign(c) {
				return
			}

			var event Event
			err := c.BindWith(&event, binding.XML)
			if err != nil {
				return
			}

			c.XML(200, handleMessage(&event))

		default:
			return
		}
	})

	authorizeGet := func(c *mel.Context, url string, result interface{}) error {
		rep, err := srv.client.Client.Get(url)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return err
		}
		defer rep.Body.Close()

		if rep.StatusCode != http.StatusOK {
			c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("http.Status: %s", rep.Status))
			return err
		}

		err = json.NewDecoder(rep.Body).Decode(result)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return err
		}

		r := result.(Error)
		if r.Code() != OK {
			err = r.(error)
			c.AbortWithError(http.StatusUnauthorized, err)
			return err
		}

		return nil
	}

	handleAuthorize := func(c *mel.Context, url, state, redirectURL string) {
		type Result struct {
			AccessToken  string `json:"access_token"`
			ExpiresIn    int64 `json:"expires_in"`
			RefreshToken string `json:"refresh_token"`
			OpenID       string `json:"openid"`
			Scope        string `json:"scope"`
			State        string `json:"state,omitempty"`
		}

		type ResultWithErr struct {
			Result
			Err
		}

		var result ResultWithErr
		err := authorizeGet(c, url, &result)
		if err != nil {
			return
		}

		result.State = state
		if redirectURL != "" {
			data, err := json.Marshal(&result.Result)
			if err != nil {
				c.AbortWithError(http.StatusUnauthorized, err)
				return
			}
			redirectURL = string(URL(redirectURL).Query("wechat", string(data)))
			c.Redirect(http.StatusMovedPermanently, redirectURL)
		} else {
			c.JSON(http.StatusOK, &result.Result)
		}
	}

	srv.Get(srv.urlPrefix+"/token", func(c *mel.Context) {
		code := c.Query("code")
		state := c.Query("state")
		redirectURL := c.Query("url")

		url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", srv.client.appId, srv.client.appSecret, code)

		handleAuthorize(c, url, state, redirectURL)
	})

	srv.Get(srv.urlPrefix+"/refresh-token", func(c *mel.Context) {
		refreshToken := c.Query("refresh_token")

		url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/refresh_token?appid=%s&grant_type=refresh_token&refresh_token=%s", srv.client.appId, refreshToken)

		handleAuthorize(c, url, "", "")
	})

	srv.Get(srv.urlPrefix+"/userinfo", func(c *mel.Context) {
		token := c.Query("access_token")
		openID := c.Query("openid")

		url := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN", token, openID)

		type Result struct {
			OpenID       string `json:"openid"`
			Nickname	string `json:"nickname"`
			Sex int `json:"sex"`
			Province string `json:"province"`
			City string `json:"city"`
			Country string `json:"country"`
			HeadImgURL string `json:"headimgurl"`
			Privilege []string `json:"privilege	"`
			UnionID string `json:"unionid"`
		}

		type ResultWithErr struct {
			Result
			Err
		}

		var result ResultWithErr
		err := authorizeGet(c, url, &result)
		if err != nil {
			return
		}
		c.JSON(http.StatusOK, &result.Result)
	})

	srv.Get(srv.urlPrefix+"/signature", func(c *mel.Context) {
		timestamp := c.Query("timestamp")
		noncestr := c.Query("noncestr")
		url := c.Query("url")
		refresh := c.Query("refresh")

		var ticket string
		var err error
		if refresh != "" && (refresh == "true" || refresh == "True" || refresh == "1") {
			ticket, err = srv.client.RefreshTicket("")
		} else {
			ticket, err = srv.client.Ticket()
		}
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		strs := sort.StringSlice{
			"timestamp=" + timestamp,
			"noncestr=" + noncestr,
			"url=" + url,
			"jsapi_ticket=" + ticket,
		}
		strs.Sort()
		h := sha1.New()
		buf := bufio.NewWriterSize(h, 1024)
		for i, s := range strs {
			buf.WriteString(s)
			if i < len(strs)-1 {
				buf.WriteByte('&')
			}
		}
		buf.Flush()
		sign := hex.EncodeToString(h.Sum(nil))
		c.JSON(http.StatusOK, map[string]string{
			"signature": sign,
		})
	})

	return srv
}

func computeSign(elements ...string) string {
	strs := sort.StringSlice(elements)
	strs.Sort()

	h := sha1.New()

	buf := bufio.NewWriterSize(h, 1024)
	for _, s := range strs {
		buf.WriteString(s)
	}
	buf.Flush()

	return hex.EncodeToString(h.Sum(nil))
}
