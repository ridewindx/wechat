package public

import (
	"fmt"
	"encoding/json"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

const (
	wechatTokenUrl   = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
	validityDuration = time.Duration(7200) * time.Second
)

type refreshResult struct {
	token string
	err   error
}

type TokenAccessor struct {
	appId     string
	appSecret string
	url string

	refreshReq chan string
	refreshRep chan refreshResult
	done       chan struct{}

	token     atomic.Value
}

func NewTokenAccessor(appId, appSecret string) (ta *TokenAccessor) {
	ta = &TokenAccessor{
		appId:      url.QueryEscape(appId),
		appSecret:  url.QueryEscape(appSecret),
		refreshReq: make(chan string),
		refreshRep: make(chan refreshResult),
		done:       make(chan struct{}),
	}
	ta.url = fmt.Sprintf(wechatTokenUrl, ta.appId, ta.appSecret)
	return
}

func (ta *TokenAccessor) Start() {
	go ta.updateTokenLoop()
}

func (ta *TokenAccessor) Stop() {
	ta.done <- struct{}{}
	<-ta.done
}

func (ta *TokenAccessor) Token() (token string, err error) {
	token, ok := ta.token.Load().(string)
	if ok && token != "" {
		return token, nil
	}

	return ta.RefreshToken("")
}

func (ta *TokenAccessor) RefreshToken(usedToken string) (token string, err error) {
	ta.refreshReq <- usedToken
	rep := <-ta.refreshRep
	return rep.token, rep.err
}

func (ta *TokenAccessor) updateTokenLoop() {
	tickDuration := validityDuration
	ticker := time.NewTicker(tickDuration)

LOOP:
	for {
		select {
		case <-ticker.C:
			_, expiresIn, err := ta.updateToken()
			if err != nil {
				break
			}
			newTickDuration := time.Duration(expiresIn) * time.Second
			if newTickDuration < tickDuration || newTickDuration-tickDuration > 5*time.Second {
				tickDuration = newTickDuration
				ticker.Stop()
				ticker = time.NewTicker(tickDuration)
			}

		case usedToken := <-ta.refreshReq:
			if usedToken != "" {
				token, ok := ta.token.Load().(string)
				if ok && token != "" && usedToken != token {
					ta.refreshRep <- refreshResult{token: token}
					break
				}
			}
			token, expiresIn, err := ta.updateToken()
			if err != nil {
				ta.refreshRep <- refreshResult{err: err}
				break
			}

			ta.refreshRep <- refreshResult{token: token}

			tickDuration = time.Duration(expiresIn) * time.Second
			ticker.Stop()
			ticker = time.NewTicker(tickDuration)

		case <-ta.done:
			break LOOP
		}
	}

	ticker.Stop()
	ta.done <- struct{}{}
}

func (ta *TokenAccessor) updateToken() (token string, expiresIn int64, err error) {
	defer ta.token.Store(token)

	rep, err := http.Get(ta.url)
	if err != nil {
		return
	}
	defer rep.Body.Close()

	if rep.StatusCode != http.StatusOK {
		err = fmt.Errorf("http.Status: %s", rep.Status)
		return
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		Err
	}

	if err = json.NewDecoder(rep.Body).Decode(&result); err != nil {
		return
	}

	if result.Code() != OK {
		err = &result.Err
		return
	}

	switch e := result.ExpiresIn; {
	case e > 60*60*24*365:
		err = fmt.Errorf("expires_in too large: %d", e)
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
		err = fmt.Errorf("expires_in too small: %d", e)
		return
	}

	token, expiresIn = result.AccessToken, result.ExpiresIn
	return
}
