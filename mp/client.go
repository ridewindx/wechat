package mp

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"fmt"
)

var BASE_URL URL = "https://api.weixin.qq.com/cgi-bin"

type Client struct {
	*TokenAccessor
	*http.Client
}

func NewClient(appId, appSecret string, needsTicket bool) *Client {
	return &Client{
		TokenAccessor: NewTokenAccessor(appId, appSecret, needsTicket),
		Client: http.DefaultClient,
	}
}

func (c *Client) call(u URL, rep interface{}, streamRep io.Writer, request func(URL) (*http.Response, error)) error {
	token, err := c.Token()
	if err != nil {
		return err
	}

	firstTime := true

RETRY:
	r, err := request(u.Query("access_token", token))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("http.Status: %s", r.Status)
	}

	if streamRep != nil {
		contentDisposition := r.Header.Get("Content-Disposition")
		contentType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if contentDisposition != "" && contentType != "text/plain" && contentType != "application/json" {
			_, err = io.Copy(streamRep, r.Body)
			return err
		}
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
		token, err = c.RefreshToken(token)
		if err != nil {
			return err
		}
		goto RETRY
	}

	return rep.(error)
}

func (c *Client) Get(u URL, rep interface{}) error {
	return c.call(u, rep, nil, func(u URL) (*http.Response, error) {
		return c.Client.Get(string(u))
	})
}

func (c *Client) Post(u URL, req, rep interface{}) error {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(req)
	if err != nil {
		return err
	}

	return c.call(u, rep, nil, func(u URL) (*http.Response, error) {
		return c.Client.Post(string(u), "application/json; charset=utf-8", buf)
	})
}

var MaxMemoryForFile = 10 * 1024 * 1024

type fileBuf struct {
	bytes.Buffer
	*os.File
	n int
}

func (fb *fileBuf) Write(p []byte) (n int, err error) {
	if fb.File != nil {
		n, err = fb.File.Write(p)
		fb.n += n
		return
	}

	pr := bytes.NewReader(p)
	nn, err := io.CopyN(&fb.Buffer, pr, int64(MaxMemoryForFile+1-fb.n))
	n = int(nn)
	fb.n += n
	if err != nil && err != io.EOF {
		return
	}
	if fb.n > MaxMemoryForFile && fb.File == nil {
		// too big, write to disk and flush buffer
		var file *os.File
		file, err = ioutil.TempFile("", "multipart-")
		if err != nil {
			return
		}
		p = p[n:]
		fb.n, err = file.Write(fb.Buffer.Bytes())
		if err == nil {
			var nn int
			nn, err = file.Write(p)
			n += nn
			fb.n += nn
		}
		if err != nil {
			file.Close()
			os.Remove(file.Name())
			return
		}

		fb.File = file
	}
	return
}

func (fb *fileBuf) Close() error {
	if fb.File != nil {
		return fb.File.Close()
	}
	return nil
}

func (c *Client) UploadFile(u URL, name, filePath string, extraFields map[string]string, rep interface{}) error {
	var buf fileBuf
	defer buf.Close()

	mp := multipart.NewWriter(&buf)
	partWriter, err := mp.CreateFormFile(name, filepath.Base(filePath))
	if err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = io.Copy(partWriter, file); err != nil {
		return err
	}

	for k, v := range extraFields {
		if err = mp.WriteField(k, v); err != nil {
			return err
		}
	}

	if err = mp.Close(); err != nil {
		return err
	}

	var reader io.ReadSeeker
	if buf.File != nil {
		reader = buf.File
	} else {
		reader = bytes.NewReader(buf.Buffer.Bytes())
	}

	return c.call(u, rep, nil, func(u URL) (*http.Response, error) {
		_, err := reader.Seek(0, 0)
		if err != nil {
			return nil, err
		}
		return c.Client.Post(string(u), mp.FormDataContentType(), reader)
	})
}

func (c *Client) DownloadFile(u URL, req interface{}, filePath string, rep interface{}) (err error) {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer func() {
		file.Close()
		if err != nil {
			os.Remove(filePath)
		}
	}()

	return c.call(u, rep, file, func(u URL) (*http.Response, error) {
		if req == nil {
			return c.Client.Get(string(u))
		} else {
			buf := &bytes.Buffer{}
			err := json.NewEncoder(buf).Encode(req)
			if err != nil {
				return nil, err
			}

			return c.Client.Post(string(u), "application/json; charset=utf-8", buf)
		}
	})
}

type URL string

func (u URL) Join(segment string) URL {
	return URL(string(u) + segment)
}

func (u URL) Query(key, value string) URL {
	buf := bytes.NewBufferString(string(u))
	if strings.IndexByte(string(u), '?') > -1 {
		buf.WriteByte('&')
	} else {
		buf.WriteByte('?')
	}
	buf.WriteString(key)
	buf.WriteByte('=')
	buf.WriteString(url.QueryEscape(value))
	return URL(buf.String())
}
