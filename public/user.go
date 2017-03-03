package public

type OpenId string

type UserList struct {
	Total int `json:"total"`
	Count int `json:"count"`

	Data struct {
		Ids []string `json:"openid,omitempty"`
	} `json:"data"`

	NextId string `json:"next_openid"`
}

func (c *client) GetUserList(nextId string) (*UserList, error) {
	u := BASE_URL.Join("/usr/get")
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
