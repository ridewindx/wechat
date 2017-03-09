package public

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
)

func (ctx *Context) ReplyTransferToCustomerService(customerAccount ...string) {
	type TransInfo struct {
		CustomerAccount string `xml:"kfAccount"`
	}

	var rep = struct {
		XMLName struct{} `xml:"xml"`
		*EventHeader
		*TransInfo `xml:"TransInfo,omitempty"`
	}{
		EventHeader: responseEventHeader("transfer_customer_service", ctx.Event),
	}

	if len(customerAccount) > 0 {
		rep.TransInfo = &TransInfo{customerAccount[0]}
	}

	ctx.WriteResponse(&rep)
	return
}

type Customer struct {
	Id           string `json:"kf_id"`
	Account      string `json:"kf_account"` // identifier@account
	Nickname     string `json:"kf_nick"`
	HeadImageURL string `json:"kf_headimgurl"`
}

type OnlineCustomer struct {
	Id                string `json:"kf_id"`
	Account           string `json:"kf_account"`    // identifier@account
	Status            int    `json:"status"`        // 1: PC Online, 2: Mobile Online, 3: PC and Mobile Online
	AutoAcceptNum     int    `json:"auto_accept"`   // maximum number of auto accepting
	AcceptedCaseCount int    `json:"accepted_case"` // count of accepted cases
}

func (c *client) GetCustomers() (customers []Customer, err error) {
	u := BASE_URL.Join("/customservice/getkflist")

	var rep struct {
		Err
		customers []Customer `json:"kf_list"`
	}

	err = c.Get(u, &rep)
	if err != nil {
		return
	}

	customers = rep.customers
	return
}

func (c *client) GetOnlineCustomers() (customers []OnlineCustomer, err error) {
	u := BASE_URL.Join("/customservice/getonlinekflist")

	var rep struct {
		Err
		Customers []OnlineCustomer `json:"kf_online_list"`
	}

	err = c.Get(u, &rep)
	if err != nil {
		return
	}

	customers = rep.Customers
	return
}

func (c *client) createOrUpdateCustomer(action, account, nickname, password string, isPlain bool) (err error) {
	u := BASE_URL.Join("/customservice/kfaccount/" + action)

	if password != "" && isPlain {
		md5Sum := md5.Sum([]byte(password))
		password = hex.EncodeToString(md5Sum[:])
	}

	var req = struct {
		Account  string `json:"kf_account"`
		Nickname string `json:"nickname,omitempty"`
		Password string `json:"password,omitempty"`
	}{
		Account:  account,
		Nickname: nickname,
		Password: password,
	}

	var rep Err

	err = c.Post(u, &req, &rep)
	return
}

func (c *client) CreateCustomer(account, nickname, password string, isPlain bool) (err error) {
	return c.createOrUpdateCustomer("add", account, nickname, password, isPlain)
}

func (c *client) UpdateCustomer(account, nickname, password string, isPlain bool) (err error) {
	return c.createOrUpdateCustomer("update", account, nickname, password, isPlain)
}

func (c *client) DeleteCustomer(account string) (err error) {
	u := BASE_URL.Join("/customservice/kfaccount/del").Query("kf_account", account)

	var rep Err

	err = c.Get(u, &rep)
	return
}
