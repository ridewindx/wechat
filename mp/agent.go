package mp

import (
	"crypto/md5"
	"encoding/hex"
	"time"
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

func (c *Client) GetAgents() (agents []Agent, err error) {
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

func (c *Client) GetOnlineAgents() (agents []OnlineAgent, err error) {
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

func (c *Client) createOrUpdateAgent(action, account, nickname, password string, isPlain bool) (err error) {
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

func (c *Client) CreateAgent(account, nickname, password string, isPlain bool) (err error) {
	return c.createOrUpdateAgent("add", account, nickname, password, isPlain)
}

func (c *Client) UpdateAgent(account, nickname, password string, isPlain bool) (err error) {
	return c.createOrUpdateAgent("update", account, nickname, password, isPlain)
}

func (c *Client) DeleteAgent(account string) (err error) {
	u := BASE_URL.Join("/customservice/kfaccount/del").Query("kf_account", account)

	var rep Err

	err = c.Get(u, &rep)
	return
}

func (c *Client) UploadAgentHeadImage(account, filePath string) (err error) {
	u := BASE_URL.Join("/customservice/kfaccount/uploadheadimg").Query("kf_account", account)

	var rep Err

	return c.UploadFile(u, "media", filePath, nil, &rep)
}

func (c *Client) createOrCloseAgentSession(action, account, openId, text string) (err error) {
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

func (c *Client) CreateAgentSession(account, openId, text string) (err error) {
	return c.createOrCloseAgentSession("create", account, openId, text)
}

func (c *Client) CloseAgentSession(account, openId, text string) (err error) {
	return c.createOrCloseAgentSession("close", account, openId, text)
}

type AgentSession struct {
	OpenId     string `json:"openid"`
	Account    string `json:"kf_account"` // accepted agent account; empty if not accepted
	CreateTime int64  `json:"createtime"`
}

func (c *Client) GetAgentSessionForCustomer(openId string) (session *AgentSession, err error) {
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

func (c *Client) GetAgentSessions(account string) (sessions []AgentSession, err error) {
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

func (c *Client) GetWaitingAgentSessions() (totalCount int, sessions []AgentSession, err error) {
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

const (
	OpCreateSession      = 1000
	OpAcceptSession      = 1001
	OpInitiateSession    = 1002
	OpSwitchSession      = 1003
	OpCloseSession       = 1004
	OpRobSession         = 1005
	OpBackendRecvMessage = 2001
	OpAgentSendMessage   = 2002
	OpAgentRecvMessage   = 2003
)

type TimeSpan struct {
	StartTime int64 `json:"starttime"` // UNIX timestamp
	EndTime   int64 `json:"endtime"`   // UNIX timestamp; EndTime-StartTime <= 24 hours
}

func NewTimeSpanAfter(start time.Time, duration time.Duration) *TimeSpan {
	var ts TimeSpan
	ts.StartTime = start.Unix()
	ts.EndTime = start.Add(duration).Unix()
	return &ts
}

func NewTimeSpanBefore(end time.Time, duration time.Duration) *TimeSpan {
	var ts TimeSpan
	ts.StartTime = end.Add(-duration).Unix()
	ts.EndTime = end.Unix()
	return &ts
}

type MsgRecord struct {
	Agent     string `json:"worker"` // agent account
	OpenId    string `json:"openid"`
	OpCode    int    `json:"opercode"`
	Timestamp int64  `json:"time"` // UNIX timestamp
	Text      string `json:"text"` // message text
}

func (c *Client) GetAgentMsgRecords(timeSpan *TimeSpan, pageIndex int, pageSize ...int) (records []MsgRecord, err error) {
	u := BASE_URL.Join("/customservice/msgrecord/getrecord")

	if pageIndex < 1 {
		panic("invalid page index")
	}

	var size = 50
	if len(pageSize) > 0 {
		size = pageSize[0]
		if size < 1 || size > 50 {
			panic("invalid page size")
		}
	}

	var req = struct {
		TimeSpan
		PageIndex int `json:"pageindex"` // base 1
		PageSize  int `json:"pagesize"`  // limit 50
	}{
		TimeSpan:  *timeSpan,
		PageIndex: pageIndex,
		PageSize:  size,
	}

	var rep struct {
		Err
		Records []MsgRecord `json:"recordlist"`
	}

	err = c.Post(u, &req, &rep)
	if err != nil {
		return
	}

	records = rep.Records
	return
}
