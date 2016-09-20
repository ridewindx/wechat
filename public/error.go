package public

import "fmt"

const (
    OK = 0
    InvalidCredential = 40001
    AccessTokenExpired = 42001
)

type Error struct {
    Code int  `json:"errcode"`
    Msg  string `json:"errmsg"`
}

func (err *Error) Error() string {
    return fmt.Sprintf("errcode: %d, errmsg: %s", err.Code, err.Msg)
}
