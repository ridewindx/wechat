package mp

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
	wechatTicketUrl  = "https://api.weixin.qq.com/cgi-bin/ticket/getticket?access_token=%s&type=jsapi"
	wechatCorpTokenUrl   = "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s"
	validityDuration = time.Duration(7200) * time.Second
)

type refreshData struct {
	token string
	ticket string
}

type refreshDataResult struct {
	refreshData
	err   error
}

type TokenAccessor struct {
	appID     string
	appSecret string

	needsTicket bool

	url       string

	refreshReq chan refreshData
	refreshRep chan refreshDataResult
	done       chan struct{}

	token  atomic.Value
	ticket atomic.Value
}

func NewTokenAccessor(appId, appSecret string, needsTicket bool) (ta *TokenAccessor) {
	ta = &TokenAccessor{
		appID:       url.QueryEscape(appId),
		appSecret:   url.QueryEscape(appSecret),
		needsTicket: needsTicket,
		refreshReq:  make(chan refreshData),
		refreshRep:  make(chan refreshDataResult),
		done:        make(chan struct{}),
	}
	ta.url = fmt.Sprintf(wechatTokenUrl, ta.appID, ta.appSecret)
	return
}


func NewCorpTokenAccessor(corpID, corpSecret string) (ta *TokenAccessor) {
	ta = NewTokenAccessor(corpID, corpSecret, false)
	ta.url = fmt.Sprintf(wechatCorpTokenUrl, ta.appID, ta.appSecret)
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
	ta.refreshReq <- refreshData{token: usedToken}
	rep := <-ta.refreshRep
	return rep.token, rep.err
}

func (ta *TokenAccessor) Ticket() (ticket string, err error) {
	ticket, ok := ta.ticket.Load().(string)
	if ok && ticket != "" {
		return ticket, nil
	}

	return ta.RefreshTicket("")
}

func (ta *TokenAccessor) RefreshTicket(usedTicket string) (ticket string, err error) {
	ta.refreshReq <- refreshData{ticket: usedTicket}
	rep := <-ta.refreshRep
	return rep.ticket, rep.err
}

func (ta *TokenAccessor) updateTokenLoop() {
	tickDuration := validityDuration
	ticker := time.NewTicker(tickDuration)

LOOP:
	for {
		select {
		case <-ticker.C:
			_, _, expiresIn, err := ta.updateToken()
			if err != nil {
				break
			}
			newTickDuration := time.Duration(expiresIn) * time.Second
			if newTickDuration < tickDuration || newTickDuration-tickDuration > 5*time.Second {
				tickDuration = newTickDuration
				ticker.Stop()
				ticker = time.NewTicker(tickDuration)
			}

		case req := <-ta.refreshReq:
			if req.token != "" {
				token, ok := ta.token.Load().(string)
				if ok && token != "" && req.token != token {
					ta.refreshRep <- refreshDataResult{refreshData: refreshData{token: token}}
					break
				}
			}
			if req.ticket != "" {
				ticket, ok := ta.token.Load().(string)
				if ok && ticket != "" && req.ticket != ticket {
					ta.refreshRep <- refreshDataResult{refreshData: refreshData{ticket: ticket}}
					break
				}
			}
			token, ticket, expiresIn, err := ta.updateToken()
			if err != nil {
				ta.refreshRep <- refreshDataResult{err: err}
				break
			}

			ta.refreshRep <- refreshDataResult{refreshData: refreshData{token, ticket}}

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

func (ta *TokenAccessor) updateToken() (token, ticket string, expiresIn int64, err error) {
	defer ta.token.Store(token)

	token, expiresIn, err = ta.update(ta.url)
	if err != nil {
		return
	}

	if !ta.needsTicket {
		return
	}

	defer ta.ticket.Store(ticket)

	var ticketExpires int64
	ticket, ticketExpires, err = ta.update(fmt.Sprintf(wechatTicketUrl, token))
	if err == nil && ticketExpires < expiresIn {
		expiresIn = ticketExpires
	}

	return
}

func (ta *TokenAccessor) update(url string) (result string, expiresIn int64, err error) {
	rep, err := http.Get(url)
	if err != nil {
		return
	}
	defer rep.Body.Close()

	if rep.StatusCode != http.StatusOK {
		err = fmt.Errorf("http.Status: %s", rep.Status)
		return
	}

	var response struct {
		AccessToken string `json:"access_token,omitempty"`
		Ticket string `json:"ticket,omitempty"`
		ExpiresIn   int64  `json:"expires_in"`
		Err
	}

	if err = json.NewDecoder(rep.Body).Decode(&response); err != nil {
		return
	}

	if response.Code() != OK {
		err = &response.Err
		return
	}

	switch e := response.ExpiresIn; {
	case e > 60*60*24*365:
		err = fmt.Errorf("expires_in too large: %d", e)
		return
	case e > 60*60:
		response.ExpiresIn -= 60 * 10
	case e > 60*30:
		response.ExpiresIn -= 60 * 5
	case e > 60*5:
		response.ExpiresIn -= 60
	case e > 60:
		response.ExpiresIn -= 10
	default:
		err = fmt.Errorf("expires_in too small: %d", e)
		return
	}

	if response.AccessToken != "" {
		result = response.AccessToken
	} else {
		result = response.Ticket
	}
	expiresIn = response.ExpiresIn
	return
}
