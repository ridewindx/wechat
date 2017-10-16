package wechat

import "crypto/subtle"

func Compare(a, b string) bool {
	if subtle.ConstantTimeEq(int32(len(a)), int32(len(b))) == 1 {
		if subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 {
			return true
		}
	}
	return false
}
