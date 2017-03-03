package public

import (
	"net/http"
	"strings"
	"net/url"
	"errors"
	"encoding/json"
	"bytes"
)

const BASE_URL = "https://api.weixin.qq.com/cgi-bin"

type client struct {
	*http.Client
}

var Client = client{
	Client: http.DefaultClient,
}

func (c *client) Token() (string, error) {
	return "", nil
}

func (c *client) RefreshToken() (string, error) {
	return "", nil
}

func (c *client) call(u URL, rep interface{}, request func(URL)(*http.Response, error)) error {
	token, err := c.Token()
	if err != nil {
		return err
	}

	firstTime := true

RETRY:
	r, err := request(u.Query("access_token", url.QueryEscape(token)))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return errors.New(r.Status)
	}

	err = json.NewDecoder(r.Body).Decode(rep)
	if err != nil {
		return err
	}

	e := rep.(Error)
	if e.Code() == OK {
		return nil
	}
	if (e.Code() == InvalidCredential || e.Code() == AccessTokenExpired) && firstTime {
		firstTime = false
		token, err = c.RefreshToken()
		if err != nil {
			return err
		}
		goto RETRY
	}

	return rep.(error)
}

func (c *client) Get(u URL, rep interface{}) error {
	return c.call(u, rep, func(u URL)(*http.Response, error) {
		return c.Client.Get(u)
	})
}

func (c *client) Post(u URL, req, rep interface{}) error {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(req)
	if err != nil {
		return err
	}

	return c.call(u, rep, func(u URL)(*http.Response, error) {
		return c.Client.Post(u, "application/json; charset=utf-8", buf)
	})
}

type URL string

func (u URL) Join(segment string) string {
	return u + segment
}

func (u URL) Query(key, value string) string {
	if strings.IndexByte(u, '?') != -1 {
		u += '&'
	} else {
		u += '?'
	}
	u += key + '=' + value
	return u
}
