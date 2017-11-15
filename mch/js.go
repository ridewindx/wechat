package mch

import (
	"time"
	"strconv"
	"crypto/md5"
)

func (client *Client) MakeJSAPIArgs(prepayID string) map[string]string {
	r := map[string]string{
		"appId": client.appID,
		"timeStamp": strconv.FormatInt(time.Now().Unix(), 10),
		"nonceStr": RandString(32),
		"package": "prepay_id="+prepayID,
		"signType": SignTypeMD5,
	}
	r["paySign"] = Sign(r, client.apiKey, md5.New())
	return r
}
