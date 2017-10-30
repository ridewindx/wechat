package mp

import (
	"net/http"
	"encoding/json"
	"fmt"
)

func oauth2Get(client *http.Client, url string, result interface{}) error {
	rep, err := client.Get(url)
	if err != nil {
		return err
	}
	defer rep.Body.Close()

	if rep.StatusCode != http.StatusOK {
		err = fmt.Errorf("http.Status: %s", rep.Status)
		return err
	}

	err = json.NewDecoder(rep.Body).Decode(result)
	if err != nil {
		return err
	}

	r := result.(Error)
	if r.Code() != OK {
		err = r.(error)
		return err
	}

	return nil
}

func oauth2GetToken(client *http.Client, url, state string) (*Oauth2Token, error) {
	type ResultWithErr struct {
		Oauth2Token
		Err
	}

	var result ResultWithErr
	err := oauth2Get(client, url, &result)
	if err != nil {
		return nil, err
	}

	result.State = state
	return &result.Oauth2Token, nil
}

type Oauth2Token struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	State        string `json:"state,omitempty"`
}

func (c *Client) Oauth2GetTokenAndRedirect(code, state, redirectURL string) (string, error) {
	token, err := c.Oauth2GetToken(code, state)
	if err != nil {
		return "", err
	}
	return c.Oauth2Redirect(token, redirectURL)
}

func (c *Client) Oauth2Redirect(token *Oauth2Token, redirectURL string) (string, error) {
	data, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	redirectURL = string(URL(redirectURL).Query("wechat", string(data)))
	return redirectURL, nil
}

func (c *Client) Oauth2GetToken(code, state string) (*Oauth2Token, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", c.appID, c.appSecret, code)

	return oauth2GetToken(c.Client, url, state)
}

func (c *Client) Oauth2RefreshToken(refreshToken string) (*Oauth2Token, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/refresh_token?appid=%s&grant_type=refresh_token&refresh_token=%s", c.appID, refreshToken)

	return oauth2GetToken(c.Client, url, "")
}

type Oauth2User struct {
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

func (c *Client) Oauth2GetUser(token, openID string) (*Oauth2User, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN", token, openID)

	type ResultWithErr struct {
		Oauth2User
		Err
	}

	var result ResultWithErr
	err := oauth2Get(c.Client, url, &result)
	if err != nil {
		return nil, err
	}
	return &result.Oauth2User, nil
}
