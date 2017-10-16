package public

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
)

type Server struct {
	*mel.Mel

	appId  string
	userId string

	tokenMutex sync.Mutex
	token      unsafe.Pointer

	aesKeyMutex sync.Mutex
	aesKey      unsafe.Pointer

	middlewares       []Handler
	messageHandlerMap map[string]Handler
	eventHandlerMap   map[string]Handler
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

func NewServer() *Server {
	srv := &Server{
		Mel:               mel.New(),
		messageHandlerMap: make(map[string]Handler),
		eventHandlerMap:   make(map[string]Handler),
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
		ToUserName string `xml:",cdata"`
		Encrypt    string `xml:",cdata"`
	}

	srv.Get("/", func(c *mel.Context) {
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
			index:    preStartIndex,
			handlers: append(srv.middlewares, handler),
			Event:    event,
		}

		ctx.Next()

		return ctx.response
	}

	srv.Post("/", func(c *mel.Context) {
		encryptType := c.Query("encrypt_type")
		signature := c.Query("signature")
		timestamp := c.Query("timestamp")
		nonce := c.Query("nonce")

		switch encryptType {
		case "aes":
			token := verifySignReturnToken(signature, timestamp, nonce)
			if token == "" {
				return
			}

			msgSign := c.Query("msg_signature")

			var obj EncryptMsg
			err := c.BindWith(&obj, binding.XML)
			if err != nil {
				return
			}

			if srv.userId != "" && !equal(obj.ToUserName, srv.userId) {
				return
			}

			computedSign := computeSign(token, timestamp, nonce, obj.Encrypt)
			if !equal(computedSign, msgSign) {
				return
			}

			encryptedMsg, err := base64.StdEncoding.DecodeString(obj.Encrypt)
			if err != nil {
				return
			}

			current, last := srv.GetAESKey()
			aesKey := current
			random, msg, appId, err := decryptMsg(encryptedMsg, []byte(aesKey))
			if err != nil {
				if last == "" {
					return
				}
				aesKey = last
				random, msg, appId, err = decryptMsg(encryptedMsg, []byte(aesKey))
				if err != nil {
					return
				}
			} else {
				srv.deleteLastAESKey()
			}
			if srv.appId != "" && string(appId) != srv.appId {
				return
			}

			var event Event
			if err = xml.Unmarshal(msg, &event); err != nil {
				return
			}

			repBytes, err := xml.Marshal(handleMessage(&event))
			if err != nil {
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

			c.XML(200, &EncryptRepMsg{encryptedRepStr, repSignature, timestamp, nonce})

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
