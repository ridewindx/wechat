package public

const (
	MsgText  = "text"
	MsgImage = "image"
	MsgVoice = "voice"
	MsgVideo = "mpvideo"
	MsgNews  = "mpnews"
	MsgCard  = "wxcard"
)

type msgFilter struct {
	isToAll bool `json:"is_to_all"`
	groupId int  `json:"group_id"`
}

type msgHeader struct {
	*msgFilter `json:"filter,omitempty"`
	msgType    string   `json:"msgtype"`
	toUsers    []string `json:"touser,omitempty"`
}

func (c *client) send(msg interface{}, byUsers bool) (id, dataId int, err error) {
	var u URL
	if byUsers {
		u = BASE_URL.Join("/message/mass/send") // send by user ids
	} else {
		u = BASE_URL.Join("/message/mass/sendall") // send all, or by group id
	}

	var rep struct {
		Err
		Id     int `json:"msg_id"`
		DataId int `json:"msg_data_id"` // only exists for MsgNews
	}

	err = c.Post(u, msg, &rep)
	if err != nil {
		return
	}

	id = rep.Id
	dataId = rep.DataId
	return
}

func newMsgHeader(msgType string, groupId []int, userIds []string) *msgHeader {
	var f *msgFilter
	if userIds == nil {
		if len(groupId) > 0 {
			f = &msgFilter{false, groupId[0]}
		} else {
			f = &msgFilter{true}
		}
	}
	return &msgHeader{
		msgFilter: f,
		msgType:   msgType,
		toUsers:   userIds,
	}
}

func (c *client) sendText(content string, groupId []int, userIds []string) (id int, err error) {
	var msg = struct {
		*msgHeader
		Text struct {
			Content string `json:"content"`
		} `json:"text"`
	}{
		msgHeader: newMsgHeader(MsgText, groupId, userIds),
		Text: {
			Content: content,
		},
	}

	id, _, err = c.send(&msg, (userIds != nil))
	return
}

func (c *client) SendTextAll(content string, groupId ...int) (id int, err error) {
	return c.sendText(content, groupId, nil)
}

func (c *client) SendTextByUsers(content string, userIds []string) (id int, err error) {
	return c.sendText(content, nil, userIds)
}

func (c *client) sendImage(mediaId string, groupId []int, userIds []string) (id int, err error) {
	var msg = struct {
		msgHeader
		Image struct {
			MediaId string `json:"media_id"`
		} `json:"image"`
	}{
		msgHeader: newMsgHeader(MsgImage, groupId, userIds),
		Image: {
			MediaId: mediaId,
		},
	}

	id, _, err = c.send(&msg, (userIds != nil))
	return
}

func (c *client) SendImageAll(mediaId string, groupId ...int) (id int, err error) {
	return c.sendImage(mediaId, groupId, nil)
}

func (c *client) SendImageByUsers(mediaId string, userIds []string) (id int, err error) {
	return c.sendImage(mediaId, nil, userIds)
}

func (c *client) sendVoice(mediaId string, groupId []int, userIds []string) (id int, err error) {
	var msg = struct {
		msgHeader
		Voice struct {
			MediaId string `json:"media_id"`
		} `json:"voice"`
	}{
		msgHeader: newMsgHeader(MsgVoice, groupId, userIds),
		Voice: {
			MediaId: mediaId,
		},
	}

	id, _, err = c.send(&msg, (userIds != nil))
	return
}

func (c *client) SendVoiceAll(mediaId string, groupId ...int) (id int, err error) {
	return c.sendVoice(mediaId, groupId, nil)
}

func (c *client) SendVoiceByUsers(mediaId string, userIds []string) (id int, err error) {
	return c.sendVoice(mediaId, nil, userIds)
}

func (c *client) sendVideo(mediaId string, groupId []int, userIds []string) (id int, err error) {
	var msg = struct {
		msgHeader
		Video struct {
			MediaId string `json:"media_id"`
		} `json:"mpvideo"`
	}{
		msgHeader: newMsgHeader(MsgVideo, groupId, userIds),
		Video: {
			MediaId: mediaId,
		},
	}

	id, _, err = c.send(&msg, (userIds != nil))
	return
}

func (c *client) SendVideoAll(mediaId string, groupId ...int) (id int, err error) {
	return c.sendVideo(mediaId, groupId, nil)
}

func (c *client) SendVideoByUsers(mediaId string, userIds []string) (id int, err error) {
	return c.sendVideo(mediaId, nil, userIds)
}

func (c *client) sendNews(mediaId string, groupId []int, userIds []string) (id, dataId int, err error) {
	var msg = struct {
		msgHeader
		News struct {
			MediaId string `json:"media_id"`
		} `json:"mpnews"`
	}{
		msgHeader: newMsgHeader(MsgNews, groupId, userIds),
		News: {
			MediaId: mediaId,
		},
	}

	id, dataId, err = c.send(&msg, (userIds != nil))
	return
}

func (c *client) SendNewsAll(mediaId string, groupId ...int) (id, dataId int, err error) {
	return c.sendNews(mediaId, groupId, nil)
}

func (c *client) SendNewsByUsers(mediaId string, userIds []string) (id, dataId int, err error) {
	return c.sendNews(mediaId, nil, userIds)
}

func (c *client) sendCard(cardId, cardExt string, groupId []int, userIds []string) (id int, err error) {
	var msg = struct {
		msgHeader
		Card struct {
			Id  string `json:"card_id"`
			Ext string `json:"card_ext,omitempty"`
		} `json:"wxcard"`
	}{
		msgHeader: newMsgHeader(MsgCard, groupId, userIds),
		Card: {
			Id:  cardId,
			Ext: cardExt,
		},
	}

	id, _, err = c.send(&msg, (userIds != nil))
	return
}

func (c *client) SendCardAll(cardId, cardExt string, groupId ...int) (id int, err error) {
	return c.sendCard(cardId, cardExt, groupId, nil)
}

func (c *client) SendCardByUsers(cardId, cardExt string, userIds []string) (id int, err error) {
	return c.sendCard(cardId, cardExt, nil, userIds)
}
