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
	"strconv"
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
		current: aesKey,
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

func NewServer() *Server {
	srv := &Server{
		Mel: mel.New(),
	}

	equal := func(a, b string) bool {
		return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
	}

	validateSign := func(c *mel.Context) bool {
		signature := c.Query("signature")
		timestamp := c.Query("timestamp")
		nonce := c.Query("nonce")

		currentToken, lastToken := srv.GetToken()
		token := currentToken

		isValid := func() bool {
			computedSignature := Sign(token, timestamp, nonce)
			return equal(signature, computedSignature)
		}

		if isValid() {
			srv.deleteLastToken()
			return true
		}

		if lastToken == "" {
			return false
		}

		token = lastToken
		return isValid()
	}

	type EncryptMsg struct {
		ToUserName string `xml:",cdata"`
		Encrypt    string `xml:",cdata"`
	}

	srv.Get("/", func(c *mel.Context) {
		if validateSign(c) {
			echostr := c.Query("echostr")
			c.Text(200, echostr)
		}
	})

	srv.Post("/", func(c *mel.Context) {
		encryptType := c.Query("encrypt_type")
		timestamp := c.Query("timestamp")
		nonce := c.Query("nonce")

		switch encryptType {
		case "aes":
			if !validateSign(c) {
				return
			}

			msgSign := c.Query("msg_signature")
			timestamp, err := strconv.ParseInt(c.Query("timestamp"), 10, 64)
			if err != nil {
				return
			}

			var obj EncryptMsg
			err = c.BindWith(&obj, binding.XML)
			if err != nil {
				return
			}

			if srv.userId != "" && !equal(obj.ToUserName, userId) {
				return
			}

			computedSign := MsgSign(token, timestamp, nonce, obj.Encrypt)
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
				random, msg, appId, err := decryptMsg(encryptedMsg, []byte(aesKey))
				if err != nil {
					return
				}
			} else {
				srv.deleteLastAESKey()
			}
			if srv.appId != "" && string(appId) != srv.appId {
				return
			}

		case "", "raw":
			if !validateSign(c) {
				return
			}

		default:
			return
		}
	})

	return srv
}

func Sign(token, timestamp, nonce string) string {
	strs := sort.StringSlice{token, timestamp, nonce}
	strs.Sort()

	buf := make([]byte, 0, len(token)+len(timestamp)+len(nonce))
	buf = append(buf, strs[0]...)
	buf = append(buf, strs[1]...)
	buf = append(buf, strs[2]...)

	hashsum := sha1.Sum(buf)
	return hex.EncodeToString(hashsum[:])
}

func MsgSign(token, timestamp, nonce, encryptedMsg string) string {
	strs := sort.StringSlice{token, timestamp, nonce, encryptedMsg}
	strs.Sort()

	h := sha1.New()

	bufw := bufio.NewWriterSize(h, 1024)
	bufw.WriteString(strs[0])
	bufw.WriteString(strs[1])
	bufw.WriteString(strs[2])
	bufw.WriteString(strs[3])
	bufw.Flush()

	hashsum := h.Sum(nil)
	return hex.EncodeToString(hashsum)
}
