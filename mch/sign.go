package mch

import (
	"crypto/md5"
	"encoding/hex"
	"crypto/sha1"
	"hash"
	"sort"
	"bufio"
	"bytes"
)

const (
	SignTypeMD5         = "MD5"
	SignTypeSHA1        = "SHA1"
	SignTypeHMAC_SHA256 = "HMAC-SHA256"
)

func Sign(params map[string]string, apiKey string, h hash.Hash) string {
	if h == nil {
		h = md5.New()
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := bufio.NewWriterSize(h, 128)
	for _, k := range keys {
		v := params[k]
		if v == "" {
			continue
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(v)
		buf.WriteByte('&')
	}
	buf.WriteString("key=")
	buf.WriteString(apiKey)
	buf.Flush()

	signature := make([]byte, hex.EncodedLen(h.Size()))
	hex.Encode(signature, h.Sum(nil))
	return string(bytes.ToUpper(signature))
}

func JSAPISign(appId, timeStamp, nonceStr, packageStr, signType string, apiKey string) string {
	var h hash.Hash
	switch signType {
	case SignTypeMD5:
		h = md5.New()
	case SignTypeSHA1:
		h = sha1.New()
	default:
		panic("unsupported signType")
	}
	buf := bufio.NewWriterSize(h, 128)

	buf.WriteString("appId=")
	buf.WriteString(appId)
	buf.WriteString("&nonceStr=")
	buf.WriteString(nonceStr)
	buf.WriteString("&package=")
	buf.WriteString(packageStr)
	buf.WriteString("&signType=")
	buf.WriteString(signType)
	buf.WriteString("&timeStamp=")
	buf.WriteString(timeStamp)
	buf.WriteString("&key=")
	buf.WriteString(apiKey)

	buf.Flush()
	signature := make([]byte, hex.EncodedLen(h.Size()))
	hex.Encode(signature, h.Sum(nil))
	return string(bytes.ToUpper(signature))
}
