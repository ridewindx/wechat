package mp

import "fmt"

type OpenId string

type UserList struct {
	Total int `json:"total"`
	Count int `json:"count"`

	Data struct {
		Ids []string `json:"openid,omitempty"`
	} `json:"data"`

	NextId string `json:"next_openid"`
}

func (c *Client) GetUserList(nextId string) (*UserList, error) {
	u := BASE_URL.Join("/user/get")
	if nextId != "" {
		u = u.Query("next_openid", nextId)
	}

	var rep struct {
		Err
		UserList
	}

	err := c.Get(u, &rep)
	if err != nil {
		return nil, err
	}

	return &rep.UserList, nil
}

func (c *Client) UpdateUserRemark(openId, remark string) error {
	u := BASE_URL.Join("/user/info/updateremark")

	var req = struct {
		OpenId string `json:"openid"`
		Remark string `json:"remark"`
	}{
		OpenId: openId,
		Remark: remark,
	}

	var rep Err
	err := c.Post(u, &req, &rep)
	return err
}

const (
	LangZhCN = "zh_CN"
	LangZhTW = "zh_TW"
	LangEN   = "en"
)

type User struct {
	IsSubscriber  int    `json:"subscribe"` // 0 represents not a subscriber and the following infos do not exist
	OpenId        string `json:"openid"`
	Nickname      string `json:"nickname"`
	Sex           int    `json:"sex"`      // 1: male, 2: female, 3: unknown
	Language      string `json:"language"` // zh_CN, zh_TW, en
	City          string `json:"city"`
	Province      string `json:"province"`
	Country       string `json:"country"`
	HeadImageURL  string `json:"headimgurl"`
	SubscribeTime int64  `json:"subscribe_time"`
	UnionId       string `json:"unionid,omitempty"` // exists only when the WeChat public account has been bound to WeChat open platform account
	Remark        string `json:"remark"`
	GroupId       int  `json:"groupid"`
}

func (c *Client) GetUser(openId string, lang ...string) (*User, error) {
	u := BASE_URL.Join("/user/info")
	u = u.Query("openid", openId)
	if len(lang) > 0 {
		u = u.Query("lang", lang[0])
	} else {
		u = u.Query("lang", LangZhCN)
	}

	var rep struct {
		Err
		User
	}
	err := c.Get(u, &rep)
	if err != nil {
		return nil, err
	}
	return &rep.User, nil
}

func (c *Client) GetUsers(openIds []string, lang ...string) ([]User, error) {
	u := BASE_URL.Join("/user/info/batchget")
	language := LangZhCN
	if len(lang) > 0 {
		language = lang[0]
	}

	type Item struct {
		OpenId   string `json:"openid"`
		Language string `json:"lang,omitempty"`
	}

	var req = struct {
		UserList []Item `json:"user_list,omitempty"`
	}{
		UserList: make([]Item, 0, len(openIds)),
	}
	for i := 0; i < len(openIds); i++ {
		req.UserList = append(req.UserList, Item{openIds[i], language})
	}

	var rep struct {
		Err
		Users []User `json:"user_info_list"`
	}

	err := c.Post(u, &req, &rep)
	if err != nil {
		return nil, err
	}

	return rep.Users, nil
}

type Group struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	UserCount int    `json:"count"`
}

func (c *Client) GetGroups() ([]Group, error) {
	u := BASE_URL.Join("/groups/get")

	var rep struct {
		Err
		Groups []Group `json:"groups"`
	}

	err := c.Get(u, &rep)
	if err != nil {
		return nil, err
	}

	return rep.Groups, nil
}

func (c *Client) GetGroupByUser(openId string) (groupId int, err error) {
	u := BASE_URL.Join("/groups/getid")

	var req = struct {
		OpenId string `json:"openid"`
	}{
		OpenId: openId,
	}

	var rep struct {
		Err
		GroupId int `json:"groupid"`
	}

	err = c.Post(u, &req, &rep)
	if err != nil {
		return
	}
	groupId = rep.GroupId
	return
}

func (c *Client) CreateGroup(name string) (*Group, error) {
	u := BASE_URL.Join("/groups/create")

	type group struct {
		Name string `json:"name"`
	}

	var req = struct {
		Group group `json:"group"`
	}{
		Group: group{name},
	}

	var rep struct {
		Err
		Group `json:"group"`
	}

	err := c.Post(u, &req, &rep)
	if err != nil {
		return nil, err
	}

	return &rep.Group, nil
}

func (c *Client) UpdateGroup(id string, name string) error {
	u := BASE_URL.Join("/groups/update")

	type group struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}

	var req = struct {
		Group group `json:"group"`
	}{
		Group: group{
			Id:   id,
			Name: name,
		},
	}

	var rep Err
	err := c.Post(u, &req, &rep)
	return err
}

func (c *Client) ChangeGroupForUser(openId string, groupId int) error {
	u := BASE_URL.Join("/groups/members/update")

	var req = struct {
		OpenId    string `json:"openid"`
		ToGroupId int    `json:"to_groupid"`
	}{
		OpenId:    openId,
		ToGroupId: groupId,
	}

	var rep Err
	err := c.Post(u, &req, &rep)
	return err
}

func (c *Client) ChangeGroupForUsers(openIds []string, groupId int) error {
	if len(openIds) > 50 {
		return fmt.Errorf("openIds num too big: %d", len(openIds))
	}

	u := BASE_URL.Join("/groups/members/update")

	var req = struct {
		OpenIdList []string `json:"openid_list"`
		ToGroupId  int      `json:"to_groupid"`
	}{
		OpenIdList: openIds,
		ToGroupId:  groupId,
	}

	var rep Err
	err := c.Post(u, &req, &rep)
	return err
}

func (c *Client) DeleteCroup(id string) error {
	u := BASE_URL.Join("/groups/delete")

	type group struct {
		Id string `json:"id"`
	}

	var req = struct {
		Group group `json:"group"`
	}{
		Group: group{
			Id: id,
		},
	}

	var rep Err
	err := c.Post(u, &req, &rep)
	return err
}
