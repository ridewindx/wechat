package mp

const (
	MsgText  = "text"
	MsgImage = "image"
	MsgVoice = "voice"
	MsgVideo = "mpvideo"
	MsgNews  = "mpnews"
	MsgCard  = "wxcard"
)

type MsgFilter struct {
	IsToAll bool `json:"is_to_all"`
	GroupId int  `json:"group_id"`
}

type msgHeader struct {
	*MsgFilter       `json:"filter,omitempty"`
	MsgType string   `json:"msgtype"`
	ToUsers []string `json:"touser,omitempty"`
}

type msgPreviewHeader struct {
	MsgType string `json:"msgtype"`
	ToUser  string `json:"touser,omitempty"`
	WxName  string `json:"towxname,omitempty"`
}

type Text struct {
	Content string `json:"content"`
}

const (
	SendAll        = iota
	SendByUsers
	SendForPreview
)

func (c *Client) send(msg interface{}, sendType int) (id, dataId int, err error) {
	var u URL
	switch sendType {
	case SendAll:
		u = BASE_URL.Join("/message/mass/sendall") // send all, or by group id
	case SendByUsers:
		u = BASE_URL.Join("/message/mass/send") // send by user ids
	case SendForPreview:
		u = BASE_URL.Join("/message/mass/preview") // send for preview
	default:
		panic("invalid sendType")
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
	var f *MsgFilter
	if userIds == nil {
		if len(groupId) > 0 {
			f = &MsgFilter{false, groupId[0]}
		} else {
			f = &MsgFilter{true, 0}
		}
	}
	return &msgHeader{
		MsgFilter: f,
		MsgType:   msgType,
		ToUsers:   userIds,
	}
}

func newMsgPreviewHeader(msgType, wxName string) *msgPreviewHeader {
	return &msgPreviewHeader{
		MsgType: msgType,
		ToUser:  wxName,
		WxName:  wxName,
	}
}

func getSendType(userIds []string) int {
	if userIds != nil {
		return SendByUsers
	} else {
		return SendAll
	}
}

func (c *Client) sendText(content string, groupId []int, userIds []string) (id int, err error) {
	var msg = struct {
		*msgHeader
		Text `json:"text"`
	}{
		msgHeader: newMsgHeader(MsgText, groupId, userIds),
		Text: Text{
			Content: content,
		},
	}

	id, _, err = c.send(&msg, getSendType(userIds))
	return
}

func (c *Client) SendTextForPreview(content, wxName string) (id int, err error) {
	var msg = struct {
		*msgPreviewHeader
		Text `json:"text"`
	}{
		msgPreviewHeader: newMsgPreviewHeader(MsgText, wxName),
		Text: Text{
			Content: content,
		},
	}

	id, _, err = c.send(&msg, SendForPreview)
	return
}

func (c *Client) SendTextAll(content string, groupId ...int) (id int, err error) {
	return c.sendText(content, groupId, nil)
}

func (c *Client) SendTextByUsers(content string, userIds []string) (id int, err error) {
	return c.sendText(content, nil, userIds)
}

type Image struct {
	MediaId string `json:"media_id"`
}

func (c *Client) sendImage(mediaId string, groupId []int, userIds []string) (id int, err error) {
	var msg = struct {
		*msgHeader
		Image Image `json:"image"`
	}{
		msgHeader: newMsgHeader(MsgImage, groupId, userIds),
		Image: Image{
			MediaId: mediaId,
		},
	}

	id, _, err = c.send(&msg, getSendType(userIds))
	return
}

func (c *Client) SendImageForPreview(mediaId, wxName string) (id int, err error) {
	var msg = struct {
		*msgPreviewHeader
		Image Image `json:"image"`
	}{
		msgPreviewHeader: newMsgPreviewHeader(MsgImage, wxName),
		Image: Image{
			MediaId: mediaId,
		},
	}

	id, _, err = c.send(&msg, SendForPreview)
	return
}

func (c *Client) SendImageAll(mediaId string, groupId ...int) (id int, err error) {
	return c.sendImage(mediaId, groupId, nil)
}

func (c *Client) SendImageByUsers(mediaId string, userIds []string) (id int, err error) {
	return c.sendImage(mediaId, nil, userIds)
}

type Voice struct {
	MediaId string `json:"media_id"`
}

func (c *Client) sendVoice(mediaId string, groupId []int, userIds []string) (id int, err error) {
	var msg = struct {
		*msgHeader
		Voice Voice `json:"voice"`
	}{
		msgHeader: newMsgHeader(MsgVoice, groupId, userIds),
		Voice: Voice{
			MediaId: mediaId,
		},
	}

	id, _, err = c.send(&msg, getSendType(userIds))
	return
}

func (c *Client) SendVoiceForPreview(mediaId, wxName string) (id int, err error) {
	var msg = struct {
		*msgPreviewHeader
		Voice Voice `json:"voice"`
	}{
		msgPreviewHeader: newMsgPreviewHeader(MsgVoice, wxName),
		Voice: Voice{
			MediaId: mediaId,
		},
	}

	id, _, err = c.send(&msg, SendForPreview)
	return
}

func (c *Client) SendVoiceAll(mediaId string, groupId ...int) (id int, err error) {
	return c.sendVoice(mediaId, groupId, nil)
}

func (c *Client) SendVoiceByUsers(mediaId string, userIds []string) (id int, err error) {
	return c.sendVoice(mediaId, nil, userIds)
}

type video struct {
	MediaId string `json:"media_id"`
}

func (c *Client) sendVideo(mediaId string, groupId []int, userIds []string) (id int, err error) {
	var msg = struct {
		*msgHeader
		Video video `json:"mpvideo"`
	}{
		msgHeader: newMsgHeader(MsgVideo, groupId, userIds),
		Video: video{
			MediaId: mediaId,
		},
	}

	id, _, err = c.send(&msg, getSendType(userIds))
	return
}

func (c *Client) SendVideoForPreview(mediaId, wxName string) (id int, err error) {
	var msg = struct {
		*msgPreviewHeader
		Video video `json:"mpvideo"`
	}{
		msgPreviewHeader: newMsgPreviewHeader(MsgVideo, wxName),
		Video: video{
			MediaId: mediaId,
		},
	}

	id, _, err = c.send(&msg, SendForPreview)
	return
}

func (c *Client) SendVideoAll(mediaId string, groupId ...int) (id int, err error) {
	return c.sendVideo(mediaId, groupId, nil)
}

func (c *Client) SendVideoByUsers(mediaId string, userIds []string) (id int, err error) {
	return c.sendVideo(mediaId, nil, userIds)
}

type news struct {
	MediaId string `json:"media_id"`
}

func (c *Client) sendNews(mediaId string, groupId []int, userIds []string) (id, dataId int, err error) {
	var msg = struct {
		*msgHeader
		News news `json:"mpnews"`
	}{
		msgHeader: newMsgHeader(MsgNews, groupId, userIds),
		News: news{
			MediaId: mediaId,
		},
	}

	id, dataId, err = c.send(&msg, getSendType(userIds))
	return
}

func (c *Client) SendNewsForPreview(mediaId, wxName string) (id, dataId int, err error) {
	var msg = struct {
		*msgPreviewHeader
		News news `json:"mpnews"`
	}{
		msgPreviewHeader: newMsgPreviewHeader(MsgNews, wxName),
		News: news{
			MediaId: mediaId,
		},
	}

	id, dataId, err = c.send(&msg, SendForPreview)
	return
}

func (c *Client) SendNewsAll(mediaId string, groupId ...int) (id, dataId int, err error) {
	return c.sendNews(mediaId, groupId, nil)
}

func (c *Client) SendNewsByUsers(mediaId string, userIds []string) (id, dataId int, err error) {
	return c.sendNews(mediaId, nil, userIds)
}

type Card struct {
	Id  string `json:"card_id"`
	Ext string `json:"card_ext,omitempty"`
}

func (c *Client) sendCard(cardId, cardExt string, groupId []int, userIds []string) (id int, err error) {
	var msg = struct {
		*msgHeader
		Card Card `json:"wxcard"`
	}{
		msgHeader: newMsgHeader(MsgCard, groupId, userIds),
		Card: Card{
			Id:  cardId,
			Ext: cardExt,
		},
	}

	id, _, err = c.send(&msg, getSendType(userIds))
	return
}

func (c *Client) SendCardForPreview(cardId, cardExt, wxName string) (id int, err error) {
	var msg = struct {
		*msgPreviewHeader
		Card Card `json:"wxcard"`
	}{
		msgPreviewHeader: newMsgPreviewHeader(MsgCard, wxName),
		Card: Card{
			Id:  cardId,
			Ext: cardExt,
		},
	}

	id, _, err = c.send(&msg, SendForPreview)
	return
}

func (c *Client) SendCardAll(cardId, cardExt string, groupId ...int) (id int, err error) {
	return c.sendCard(cardId, cardExt, groupId, nil)
}

func (c *Client) SendCardByUsers(cardId, cardExt string, userIds []string) (id int, err error) {
	return c.sendCard(cardId, cardExt, nil, userIds)
}

// DeleteMsg deletes the mass message(only MsgNews and MsgVideo) which was sent in half an hour.
// It invalidates the message content page, but still retains the message card.
func (c *Client) DeleteMsg(msgId int64) error {
	u := BASE_URL.Join("/message/mass/delete")

	var req = struct {
		Id int64 `json:"msg_id"`
	}{
		Id: msgId,
	}

	var rep Err

	return c.Post(u, &req, &rep)
}

func (c *Client) IsMsgSent(msgId int64) (bool, error) {
	u := BASE_URL.Join("/message/mass/get")

	var req = struct {
		Id int64 `json:"msg_id"`
	}{
		Id: msgId,
	}

	var rep struct {
		Err
		Id     int    `json:"msg_id"`
		Status string `json:"msg_status"`
	}

	err := c.Post(u, &req, &rep)
	if err != nil {
		return false, err
	}

	return rep.Status == "SEND_SUCCESS", nil
}

func (ctx *Context) ReplyText(content string) {
	var rep = struct {
		XMLName struct{} `xml:"xml"`
		*EventHeader
		Content string   `xml:"Content"`
	}{
		EventHeader: responseEventHeader("text", ctx.Event),
		Content:     content,
	}

	ctx.WriteResponse(&rep)
	return
}

func (ctx *Context) ReplyImage(mediaId string) {
	var rep = struct {
		XMLName struct{} `xml:"xml"`
		*EventHeader
		Image struct {
			MediaId string `xml:"MediaId"`
		} `xml:"Image"`
	}{
		EventHeader: responseEventHeader("image", ctx.Event),
	}

	rep.Image.MediaId = mediaId

	ctx.WriteResponse(&rep)
	return
}

func (ctx *Context) ReplyVoice(mediaId string) {
	var rep = struct {
		XMLName struct{} `xml:"xml"`
		*EventHeader
		Voice struct {
			MediaId string `xml:"MediaId"`
		} `xml:"Voice"`
	}{
		EventHeader: responseEventHeader("voice", ctx.Event),
	}

	rep.Voice.MediaId = mediaId

	ctx.WriteResponse(&rep)
	return
}

func (ctx *Context) ReplyVideo(mediaId, title, description string) {
	type Video struct {
		MediaId     string `xml:"MediaId"`
		Title       string `xml:"Title"`
		Description string `xml:"Description"`
	}

	var rep = struct {
		XMLName struct{} `xml:"xml"`
		*EventHeader
		Video            `xml:"Video"`
	}{
		EventHeader: responseEventHeader("video", ctx.Event),
		Video: Video{
			MediaId:     mediaId,
			Title:       title,
			Description: description,
		},
	}

	ctx.WriteResponse(&rep)
	return
}

type Music struct {
	Title       string `xml:"Title,omitempty"`
	Description string `xml:"Description,omitempty"`
	URL         string `xml:"MusicUrl"`
	HQURL       string `xml:"HQMusicUrl"`
	ThumbId     string `xml:"ThumbMediaId"`
}

func (ctx *Context) ReplyMusic(music *Music) {
	var rep = struct {
		XMLName struct{} `xml:"xml"`
		*EventHeader
		*Music           `xml:"Music"`
	}{
		EventHeader: responseEventHeader("music", ctx.Event),
		Music:       music,
	}

	ctx.WriteResponse(&rep)
	return
}

type ResponseArticle struct {
	Title       string `xml:"Title"`
	Description string `xml:"Description"`
	PicURL      string `xml:"PicUrl"`
	URL         string `xml:"Url"`
}

func (ctx *Context) ReplyNews(articles []ResponseArticle) {
	var rep = struct {
		XMLName  struct{}          `xml:"xml"`
		*EventHeader
		Count    int               `xml:"ArticleCount"`
		Articles []ResponseArticle `xml:"Articles>item"`
	}{
		EventHeader: responseEventHeader("news", ctx.Event),
		Count:       len(articles),
		Articles:    articles,
	}

	ctx.WriteResponse(&rep)
	return
}
