package mp

type TemplateMsgValue struct {
	Value string `json:"value"`
	Color string `json:"color,omitempty"`
}

type TemplateMsg struct {
	ToUser string `json:"touser"`
	TemplateID string `json:"template_id"`
	URL string `json:"url,omitempty"`
	MiniProgram struct{
		AppID string `json:"appid"`
		PagePath string `json:"pagepath"`
	} `json:"miniprogram,omitempty"`
	Data map[string]TemplateMsgValue `json:"data"`
}

func (c *Client) SendTemplateMessage(msg *TemplateMsg, values ...map[string]string) (msgID int64, err error) {
	if msg.Data == nil && len(values) > 0 {
		msg.Data = make(map[string]TemplateMsgValue)
		for k, v := range values[0] {
			msg.Data[k] = TemplateMsgValue{Value: v}
		}
	}

	var rep struct {
		Err
		MsgID int64 `json:"msgid"`
	}

	u := BASE_URL.Join("/message/template/send")
	err = c.Post(u, msg, &rep)
	if err != nil {
		return
	}

	msgID = rep.MsgID
	return
}
