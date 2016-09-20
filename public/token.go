package public

import (
	"errors"
	"fmt"
	"github.com/chanxuehong/wechat.v2/json"
	"math"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

const (
	wechatTokenUrl   = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
	validityDuration = time.Duration(7200) * time.Second
)

type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type refreshResult struct {
	token string
	err   error
}

type TokenAccessor struct {
	appId     string
	appSecret string

	refreshRequest  chan struct{}
	refreshResponse chan refreshResult
	done            chan struct{}

	token     atomic.Value
	lastToken string
}

func NewTokenAccessor(appId, appSecret string) (ta *TokenAccessor) {
	ta = &TokenAccessor{
		appId:           url.QueryEscape(appId),
		appSecret:       url.QueryEscape(appSecret),
		refreshRequest:  make(chan struct{}),
		refreshResponse: make(chan refreshResult),
		done:            make(chan struct{}),
	}
	return
}

func (ta *TokenAccessor) Run() {
	go ta.updateTokenLoop()
}

func (ta *TokenAccessor) Stop() {
	ta.done <- struct{}{}
	<-ta.done
}

func (ta *TokenAccessor) Token() (token string, err error) {
	t := ta.token.Load().(Token)
	if t.AccessToken != "" {
		ta.lastToken = t.AccessToken
		return t.AccessToken, nil
	}

	return ta.RefreshToken()
}

func (ta *TokenAccessor) RefreshToken() (token string, err error) {
	ta.refreshRequest <- struct{}{}
	rep := <-ta.refreshResponse
	ta.lastToken = rep.token
	return rep.token, rep.err
}

func (ta *TokenAccessor) updateTokenLoop() {
	tickDuration := validityDuration
	ticker := time.NewTicker(tickDuration)
	for {
		select {
		case <-ticker.C:
			token, err := ta.updateToken()
			if err != nil {
				break
			}
			newTickDuration := time.Duration(token.ExpiresIn) * time.Second
			if newTickDuration < tickDuration || newTickDuration-tickDuration > 5*time.Second {
				tickDuration = newTickDuration
				ticker.Stop()
				ticker = time.NewTicker(tickDuration)
			}

		case <-ta.refreshRequest:
			token, err := ta.updateToken()
			if err != nil {
				ta.refreshResponse <- refreshResult{err: err}
				break
			}

			ta.refreshResponse <- refreshResult{token: token.AccessToken}

			tickDuration = time.Duration(token.ExpiresIn) * time.Second
			ticker.Stop()
			ticker = time.NewTicker(tickDuration)

		case <-ta.done:
			<-ta.done
		}
	}
}

func (ta *TokenAccessor) updateToken() (token Token, err error) {
	if ta.lastToken != "" {
		t := ta.token.Load().(Token)
		if t.AccessToken != "" && ta.lastToken != t.AccessToken {
			return t.AccessToken, nil
		}
	}

	rep, err := http.Get(fmt.Sprintf(wechatTokenUrl, ta.appId, ta.appSecret))
	if err != nil {
		ta.token.Store(Token{})
		return
	}
	defer rep.Body.Close()

	if rep.StatusCode != http.StatusOK {
		ta.token.Store(Token{})
		err = fmt.Errorf("http.Status: %s", rep.Status)
		return
	}

	var result struct {
		Token
		Error
	}

	if err = json.NewDecoder(rep).Decode(&result); err != nil {
		ta.token.Store(Token{})
		return
	}

	if result.Code != OK {
		ta.token.Store(Token{})
		err = &result.Error
		return
	}

	switch e := result.ExpiresIn; e {
	case e > 60*60*24*365:
		ta.token.Store(Token{})
		err = errors.New(fmt.Sprintf("expires_in too large: %d", e))
		return
	case e > 60*60:
		result.ExpiresIn -= 60 * 10
	case e > 60*30:
		result.ExpiresIn -= 60 * 5
	case e > 60*5:
		result.ExpiresIn -= 60
	case e > 60:
		result.ExpiresIn -= 10
	default:
		ta.token.Store(Token{})
		err = errors.New(fmt.Sprintf("expires_in too small: %d", e))
		return
	}

	ta.token.Store(result.Token)
	return result.Token, nil
}
