package public

import (
	"crypto/md5"
	"encoding/hex"
)

func (ctx *Context) ReplyTransferToAgent(agentAccount ...string) {
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

	if len(agentAccount) > 0 {
		rep.TransInfo = &TransInfo{agentAccount[0]}
	}

	ctx.WriteResponse(&rep)
	return
}

// Customer Service Agent
type Agent struct {
	Id           string `json:"kf_id"`
	Account      string `json:"kf_account"` // identifier@account
	Nickname     string `json:"kf_nick"`
	HeadImageURL string `json:"kf_headimgurl"`
}

// Online Customer Service Agent
type OnlineAgent struct {
	Id                string `json:"kf_id"`
	Account           string `json:"kf_account"`    // identifier@account
	Status            int    `json:"status"`        // 1: PC Online, 2: Mobile Online, 3: PC and Mobile Online
	AutoAcceptNum     int    `json:"auto_accept"`   // maximum number of auto accepting
	AcceptedCaseCount int    `json:"accepted_case"` // count of accepted cases
}

func (c *client) GetAgents() (agents []Agent, err error) {
	u := BASE_URL.Join("/customservice/getkflist")

	var rep struct {
		Err
		Agents []Agent `json:"kf_list"`
	}

	err = c.Get(u, &rep)
	if err != nil {
		return
	}

	agents = rep.Agents
	return
}

func (c *client) GetOnlineAgents() (agents []OnlineAgent, err error) {
	u := BASE_URL.Join("/customservice/getonlinekflist")

	var rep struct {
		Err
		Agents []OnlineAgent `json:"kf_online_list"`
	}

	err = c.Get(u, &rep)
	if err != nil {
		return
	}

	agents = rep.Agents
	return
}

func (c *client) createOrUpdateAgent(action, account, nickname, password string, isPlain bool) (err error) {
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

func (c *client) CreateAgent(account, nickname, password string, isPlain bool) (err error) {
	return c.createOrUpdateAgent("add", account, nickname, password, isPlain)
}

func (c *client) UpdateAgent(account, nickname, password string, isPlain bool) (err error) {
	return c.createOrUpdateAgent("update", account, nickname, password, isPlain)
}

func (c *client) DeleteAgent(account string) (err error) {
	u := BASE_URL.Join("/customservice/kfaccount/del").Query("kf_account", account)

	var rep Err

	err = c.Get(u, &rep)
	return
}

func (c *client) UploadAgentHeadImage(account, filePath string) (err error) {
	u := BASE_URL.Join("/customservice/kfaccount/uploadheadimg").Query("kf_account", account)

	var rep Err

	return c.UploadFile(u, "media", filePath, nil, &rep)
}

func (c *client) createOrCloseAgentSession(action, account, openId, text string) (err error) {
	u := BASE_URL.Join("/customservice/kfsession/" + action)

	var req = struct {
		Account string `json:"kf_account"`
		OpenId  string `json:"openid"`
		Text    string `json:"text,omitempty"`
	}{
		Account: account,
		OpenId:  openId,
		Text:    text,
	}

	var rep Err

	err = c.Post(u, &req, &rep)
	return
}

func (c *client) CreateAgentSession(account, openId, text string) (err error) {
	return c.createOrCloseAgentSession("create", account, openId, text)
}

func (c *client) CloseAgentSession(account, openId, text string) (err error) {
	return c.createOrCloseAgentSession("close", account, openId, text)
}

type AgentSession struct {
	OpenId     string `json:"openid"`
	Account    string `json:"kf_account"` // accepted agent account; empty if not accepted
	CreateTime int64  `json:"createtime"`
}

func (c *client) GetAgentSessionForCustomer(openId string) (session *AgentSession, err error) {
	u := BASE_URL.Join("/customservice/kfsession/getsession").Query("openid", openId)

	var rep struct {
		Err
		AgentSession
	}

	err = c.Get(u, &rep)
	if err != nil {
		return
	}

	session = &rep.AgentSession
	session.OpenId = openId
	return
}

func (c *client) GetAgentSessions(account string) (sessions []AgentSession, err error) {
	u := BASE_URL.Join("/customservice/kfsession/getsessionlist").Query("kf_account", account)

	var rep struct {
		Err
		Sessions []AgentSession `json:"sessionlist"`
	}

	err = c.Get(u, &rep)
	if err != nil {
		return
	}

	sessions = rep.Sessions
	for i := 0; i < len(rep.Sessions); i++ {
		rep.Sessions[i].Account = account
	}
	return
}

func (c *client) GetWaitingAgentSessions() (totalCount int, sessions []AgentSession, err error) {
	u := BASE_URL.Join("/customservice/kfsession/getwaitcase")

	var rep struct {
		TotalCount int            `json:"count"`
		Sessions   []AgentSession `json:"waitcaselist,omitempty"`
	}

	err = c.Get(u, &rep)
	if err != nil {
		return
	}

	totalCount = rep.TotalCount
	sessions = rep.Sessions
	return
}
