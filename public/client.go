package public

import (
	"net/http"
	"strings"
	"net/url"
	"errors"
	"encoding/json"
	"bytes"
	"io"
	"mime/multipart"
	"io/ioutil"
	"os"
)

var BASE_URL URL = "https://api.weixin.qq.com/cgi-bin"

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
	r, err := request(u.Query("access_token", token))
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

var MaxMemoryForFile = 10*1024*1024

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
	n, err = io.CopyN(&fb.Buffer, pr, MaxMemoryForFile+1-fb.n)
	fb.n += n
	if err != nil && err != io.EOF {
		return
	}
	if fb.n > MaxMemoryForFile {
		// too big, write to disk and flush buffer
		file, err := ioutil.TempFile("", "multipart-")
		if err != nil {
			return
		}
		fb.n, err = file.Write(p)
		if cerr := file.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			os.Remove(file.Name())
			return
		}
		file.Seek(0, 0)

		fb.File = file
	}
	return
}

func (fb *fileBuf) Close() error {
	if fb.File {
		return fb.File.Close()
	}
	return nil
}

func (c *client) UploadFile(u URL, name, fileName string, file io.Reader, rep interface{}) error {
	var buf fileBuf
	defer buf.Close()

	mp := multipart.NewWriter(buf)
	partWriter, err := mp.CreateFormFile(name, fileName)
	if err != nil {
		return err
	}
	if _, err = io.Copy(partWriter, file); err != nil {
		return err
	}
	if err = mp.Close(); err != nil {
		return err
	}

	var reader io.ReadSeeker
	if buf.File {
		reader = buf.File
	} else {
		reader = bytes.NewReader(buf.Buffer.Bytes())
	}

	return c.call(u, rep, func(u URL)(*http.Response, error) {
		reader.Seek(0, 0)
		return c.Client.Post(u, mp.FormDataContentType, reader)
	})
}

type URL string

func (u URL) Join(segment string) URL {
	return u + segment
}

func (u URL) Query(key, value string) URL {
	if strings.IndexByte(u, '?') != -1 {
		u += '&'
	} else {
		u += '?'
	}
	u += key + '=' + url.QueryEscape(value)
	return u
}
