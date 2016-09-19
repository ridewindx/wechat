package wechat

import "fmt"

const (
    OK = 0
    InvalidCredential = 40001
    AccessTokenExpired = 42001
)

type Error struct {
    code int  `json:"errcode"`
    msg string `json:"errmsg"`
}

func (err *Error) Error() string {
    return fmt.Sprintf("errcode: %d, errmsg: %s", err.code, err.msg)
}
