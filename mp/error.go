package public

import "fmt"

const (
    OK = 0
    InvalidCredential = 40001
    AccessTokenExpired = 42001
)

type Error interface {
    error
    Code() int
    Msg() string
}

var _ Error = &Err{}
var _ error = &Err{}

type Err struct {
    ErrCode int  `json:"errcode"`
    ErrMsg  string `json:"errmsg"`
}

func (err *Err) Error() string {
    return fmt.Sprintf("errcode: %d, errmsg: %s", err.ErrCode, err.ErrMsg)
}

func (err *Err) Code() int {
    return err.ErrCode
}

func (err *Err) Msg() string {
    return err.ErrMsg
}
