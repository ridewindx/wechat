package public

import (
	"github.com/ridewindx/mel"
	"sync"
	"unsafe"
	"sync/atomic"
	"encoding/base64"
	"bytes"
	"sort"
	"crypto/sha1"
	"encoding/hex"
	"crypto/subtle"
	"strconv"
)

type Server struct {
	*mel.Mel

	tokenMutex sync.Mutex
	token unsafe.Pointer

	aesKeyMutex sync.Mutex
	aesKey unsafe.Pointer
}

type Token struct {
	current string
	last string
}

type AESKey struct {
	current []byte
	last []byte
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
		last: current,
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
	if bytes.Equal(aesKey, current) {
		return
	}

	k := AESKey{
		current: aesKey,
		last: current,
	}
	atomic.StorePointer(&srv.aesKey, unsafe.Pointer(&k))
}

func (srv *Server) deleteAESKey() {
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

	validateSign := func(c *mel.Context) bool {
		signature := c.Query("signature")
		timestamp := c.Query("timestamp")
		nonce := c.Query("nonce")

		currentToken, lastToken := srv.GetToken()
		token := currentToken

		isValid := func() bool {
			computedSignature := Sign(token, timestamp, nonce)
			r := subtle.ConstantTimeCompare([]byte(signature), []byte(computedSignature))
			return r == 1
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

	srv.Get("/", func(c *mel.Context) {
		if validateSign(c) {
			echostr := c.Query("echostr")
			c.Text(200, echostr)
		}
	})

	srv.Post("/", func(c *mel.Context) {
		encryptType := c.Query("encrypt_type")
		switch encryptType {
		case "aes":
			if !validateSign(c) {
				return
			}

			msgSignature := c.Query("msg_signature")
			timestamp, err := strconv.ParseInt(c.Query("timestamp"), 10, 64)
			if err != nil {
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
