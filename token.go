package wechat

import (
    "net/http"
    "fmt"
    "github.com/chanxuehong/wechat.v2/json"
    "sync/atomic"
    "time"
    "math"
    "errors"
)

const (
    wechatTokenUrl = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
    validityDuration = time.Duration(7200) * time.Second
)

type Token struct {
    accessToken string `json:"access_token"`
    expiresIn int64 `json:"expires_in"`
}

type refreshResult struct {
    token string
    err error
}

type TokenAccessor struct {
    appId string
    appSecret string

    token atomic.Value
    lastToken string

    refreshRequest chan struct{}
    refreshResponse chan refreshResult
}

func (ta *TokenAccessor) Token() (token string, err error) {
    t := ta.token.Load().(Token)
    if t.accessToken != "" {
        ta.lastToken = t.accessToken
        return t.accessToken, nil
    }

    return ta.RefreshToken()
}

func (ta *TokenAccessor) RefreshToken() (token string, err error) {
    ta.refreshRequest <- struct{}{}
    rep := <-ta.refreshResponse
    ta.lastToken = rep.token
    return rep.token, rep.err
}

func (ta *TokenAccessor) updateTokenDaemon() {
    tickDuration := validityDuration
    ticker := time.NewTicker(tickDuration)
    for {
        select {
        case <-ticker.C:
            token, err := ta.updateToken()
            if err != nil {
                break
            }
            newTickDuration := time.Duration(token.expiresIn) * time.Second
            if math.Abs(float64(newTickDuration - tickDuration)) > float64(5 * time.Second) {
                tickDuration = newTickDuration
                ticker.Stop()
                ticker = time.NewTicker(tickDuration)
            }

        case <-ta.refreshRequest:
            token, err := ta.updateToken()
            if err != nil {
                ta.refreshResponse <- refreshResult{ err: err }
                break
            }

            ta.refreshResponse <- refreshResult{ token: token.accessToken }

            tickDuration = time.Duration(token.expiresIn) * time.Second
            ticker.Stop()
            ticker = time.NewTicker(tickDuration)
        }
    }
}

func (ta *TokenAccessor) updateToken() (token Token, err error) {
    if ta.lastToken != "" {
        t := ta.token.Load().(Token)
        if t.accessToken != "" && ta.lastToken != t.accessToken {
            return t.accessToken, nil
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

    if result.code != OK {
        ta.token.Store(Token{})
        err = &result.Error
        return
    }

    switch e := result.expiresIn {
    case e > 60 * 60 * 24 * 365:
        ta.token.Store(Token{})
        err = errors.New(fmt.Sprintf("expires_in too large: %d", e))
        return
    case e > 60 * 60:
        result.expiresIn -= 60 * 10
    case e > 60 * 30:
        result.expiresIn -= 60 * 5
    case e > 60 * 5:
        result.expiresIn -= 60
    case e > 60:
        result.expiresIn -= 10
    default:
        ta.token.Store(Token{})
        err = errors.New(fmt.Sprintf("expires_in too small: %d", e))
        return
    }

    ta.token.Store(result.Token)
    return result.Token, nil
}
