package mp

import (
	"strings"
)

const (
	MsgText  = "text"
	MsgImage = "image"
	MsgVoice = "voice"
	MsgMPVideo = "mpvideo"
	MsgMPNews  = "mpnews"
	MsgCard  = "wxcard"
	MsgVideo = "video" // corp
	MsgFile = "file" // corp
	MsgNews  = "news" // corp
)

type MsgFilter struct {
	IsToAll bool `json:"is_to_all,omitempty"`
	TagID int  `json:"tag_id,omitempty"`
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

type corpMsgHeader struct {
	ToUsers string `json:"touser,omitempty"`
	ToParties string `json:"toparty,omitempty"`
	ToTags string `json:"totag,omitempty"`
	MsgType string   `json:"msgtype"`
	AgentID int64 `json:"agentid"`
	Safe bool `json:"safe"`
}

func join(slice []string) string {
	return strings.Join(slice, "|")
}

func split(s string) []string {
	return strings.Split(s, "|")
}

const (
	SendAll        = iota
	SendByUsers
	SendForPreview

	SendByParties // corp
	SendByTags // corp
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
		DataId int `json:"msg_data_id"` // only exists for MsgMPNews
	}

	err = c.Post(u, msg, &rep)
	if err != nil {
		return
	}

	id = rep.Id
	dataId = rep.DataId
	return
}

func (c *Client) corpSend(msg interface{}) error {
	var rep CorpErr
	err := c.Post(URL("https://qyapi.weixin.qq.com/cgi-bin/message/send"), msg, &rep)
	if err != nil {
		return err
	}
	return &rep
}

func newMsgHeader(msgType string, tagID []int, userIds []string) *msgHeader {
	var f *MsgFilter
	if userIds == nil {
		if len(tagID) > 0 {
			f = &MsgFilter{false, tagID[0]}
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

func (c *Client) newCorpMsgHeader(msgType string, userIDs, partyIDs, tagIDs []string) *corpMsgHeader {
	return &corpMsgHeader{join(userIDs), join(partyIDs), join(tagIDs), msgType, c.AgentID, false}
}

func getSendType(userIds []string) int {
	if userIds != nil {
		return SendByUsers
	} else {
		return SendAll
	}
}

type Text struct {
	Content string `json:"content"`
}

func (c *Client) sendText(content string, tagID []int, userIds []string) (id int, err error) {
	var msg = struct {
		*msgHeader
		Text `json:"text"`
	}{
		msgHeader: newMsgHeader(MsgText, tagID, userIds),
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

func (c *Client) SendTextAll(content string, tagID ...int) (id int, err error) {
	return c.sendText(content, tagID, nil)
}

func (c *Client) SendTextByUsers(content string, userIds []string) (id int, err error) {
	return c.sendText(content, nil, userIds)
}

func (c *Client) corpSendText(content string, userIDs, partyIDs, tagIDs []string) error {
	var msg = struct{
		*corpMsgHeader
		Text `json:"text"`
	}{
		corpMsgHeader: c.newCorpMsgHeader(MsgText, userIDs, partyIDs, tagIDs),
		Text: Text{
			Content: content,
		},
	}
	return c.corpSend(&msg)
}

func (c *Client) CorpSendTextToAll(content string) error {
	return c.corpSendText(content, []string{"@all"}, nil, nil)
}

func (c *Client) CorpSendTextByUsers(content string, userIDs []string) error {
	return c.corpSendText(content, userIDs, nil, nil)
}

func (c *Client) CorpSendTextByParties(content string, partyIDs []string) error {
	return c.corpSendText(content, nil, partyIDs, nil)
}

func (c *Client) CorpSendTextByTags(content string, tagIDs []string) error {
	return c.corpSendText(content, nil, nil, tagIDs)
}

type Image struct {
	MediaId string `json:"media_id"`
}

func (c *Client) sendImage(mediaId string, tagID []int, userIds []string) (id int, err error) {
	var msg = struct {
		*msgHeader
		Image Image `json:"image"`
	}{
		msgHeader: newMsgHeader(MsgImage, tagID, userIds),
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

func (c *Client) SendImageAll(mediaId string, tagID ...int) (id int, err error) {
	return c.sendImage(mediaId, tagID, nil)
}

func (c *Client) SendImageByUsers(mediaId string, userIds []string) (id int, err error) {
	return c.sendImage(mediaId, nil, userIds)
}

func (c *Client) corpSendImage(mediaID string, userIDs, partyIDs, tagIDs []string) error {
	var msg = struct{
		*corpMsgHeader
		Image Image `json:"image"`
	}{
		corpMsgHeader: c.newCorpMsgHeader(MsgImage, userIDs, partyIDs, tagIDs),
		Image: Image{
			MediaId: mediaID,
		},
	}
	return c.corpSend(&msg)
}

func (c *Client) CorpSendImageToAll(mediaID string) error {
	return c.corpSendImage(mediaID, []string{"@all"}, nil, nil)
}

func (c *Client) CorpSendImageByUsers(mediaID string, userIDs []string) error {
	return c.corpSendImage(mediaID, userIDs, nil, nil)
}

func (c *Client) CorpSendImageByParties(mediaID string, partyIDs []string) error {
	return c.corpSendImage(mediaID, nil, partyIDs, nil)
}

func (c *Client) CorpSendImageByTags(mediaID string, tagIDs []string) error {
	return c.corpSendImage(mediaID, nil, nil, tagIDs)
}

type Voice struct {
	MediaId string `json:"media_id"`
}

func (c *Client) sendVoice(mediaId string, tagID []int, userIds []string) (id int, err error) {
	var msg = struct {
		*msgHeader
		Voice Voice `json:"voice"`
	}{
		msgHeader: newMsgHeader(MsgVoice, tagID, userIds),
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

func (c *Client) SendVoiceAll(mediaId string, tagID ...int) (id int, err error) {
	return c.sendVoice(mediaId, tagID, nil)
}

func (c *Client) SendVoiceByUsers(mediaId string, userIds []string) (id int, err error) {
	return c.sendVoice(mediaId, nil, userIds)
}

func (c *Client) corpSendVoice(mediaID string, userIDs, partyIDs, tagIDs []string) error {
	var msg = struct{
		*corpMsgHeader
		Voice Voice `json:"voice"`
	}{
		corpMsgHeader: c.newCorpMsgHeader(MsgVoice, userIDs, partyIDs, tagIDs),
		Voice: Voice{
			MediaId: mediaID,
		},
	}
	return c.corpSend(&msg)
}

func (c *Client) CorpSendVoiceToAll(mediaID string) error {
	return c.corpSendVoice(mediaID, []string{"@all"}, nil, nil)
}

func (c *Client) CorpSendVoiceByUsers(mediaID string, userIDs []string) error {
	return c.corpSendVoice(mediaID, userIDs, nil, nil)
}

func (c *Client) CorpSendVoiceByParties(mediaID string, partyIDs []string) error {
	return c.corpSendVoice(mediaID, nil, partyIDs, nil)
}

func (c *Client) CorpSendVoiceByTags(mediaID string, tagIDs []string) error {
	return c.corpSendVoice(mediaID, nil, nil, tagIDs)
}

type MPVideo struct {
	MediaId string `json:"media_id"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

func (c *Client) sendVideo(video MPVideo, tagID []int, userIds []string) (id int, err error) {
	var msg = struct {
		*msgHeader
		Video MPVideo `json:"mpvideo"`
	}{
		msgHeader: newMsgHeader(MsgMPVideo, tagID, userIds),
		Video: video,
	}

	id, _, err = c.send(&msg, getSendType(userIds))
	return
}

func (c *Client) SendVideoForPreview(video MPVideo, wxName string) (id int, err error) {
	var msg = struct {
		*msgPreviewHeader
		Video MPVideo `json:"mpvideo"`
	}{
		msgPreviewHeader: newMsgPreviewHeader(MsgMPVideo, wxName),
		Video: video,
	}

	id, _, err = c.send(&msg, SendForPreview)
	return
}

func (c *Client) SendVideoAll(video MPVideo, tagID ...int) (id int, err error) {
	return c.sendVideo(video, tagID, nil)
}

func (c *Client) SendVideoByUsers(video MPVideo, userIds []string) (id int, err error) {
	return c.sendVideo(video, nil, userIds)
}

func (c *Client) corpSendVideo(video MPVideo, userIDs, partyIDs, tagIDs []string) error {
	var msg = struct{
		*corpMsgHeader
		Video MPVideo `json:"video"`
	}{
		corpMsgHeader: c.newCorpMsgHeader(MsgVideo, userIDs, partyIDs, tagIDs),
		Video: video,
	}
	return c.corpSend(&msg)
}

func (c *Client) CorpSendVideoToAll(video MPVideo) error {
	return c.corpSendVideo(video, []string{"@all"}, nil, nil)
}

func (c *Client) CorpSendVideoByUsers(video MPVideo, userIDs []string) error {
	return c.corpSendVideo(video, userIDs, nil, nil)
}

func (c *Client) CorpSendVideoByParties(video MPVideo, partyIDs []string) error {
	return c.corpSendVideo(video, nil, partyIDs, nil)
}

func (c *Client) CorpSendVideoByTags(video MPVideo, tagIDs []string) error {
	return c.corpSendVideo(video, nil, nil, tagIDs)
}

type File struct {
	MediaId string `json:"media_id"`
}

func (c *Client) corpSendFile(mediaID string, userIDs, partyIDs, tagIDs []string) error {
	var msg = struct{
		*corpMsgHeader
		File File `json:"file"`
	}{
		corpMsgHeader: c.newCorpMsgHeader(MsgFile, userIDs, partyIDs, tagIDs),
		File: File{mediaID},
	}
	return c.corpSend(&msg)
}

func (c *Client) CorpSendFileToAll(mediaID string) error {
	return c.corpSendFile(mediaID, []string{"@all"}, nil, nil)
}

func (c *Client) CorpSendFileByUsers(mediaID string, userIDs []string) error {
	return c.corpSendFile(mediaID, userIDs, nil, nil)
}

func (c *Client) CorpSendFileByParties(mediaID string, partyIDs []string) error {
	return c.corpSendFile(mediaID, nil, partyIDs, nil)
}

func (c *Client) CorpSendFileByTags(mediaID string, tagIDs []string) error {
	return c.corpSendFile(mediaID, nil, nil, tagIDs)
}

type news struct {
	MediaId string `json:"media_id"`
}

func (c *Client) sendNews(mediaId string, tagID []int, userIds []string) (id, dataId int, err error) {
	var msg = struct {
		*msgHeader
		News news `json:"mpnews"`
	}{
		msgHeader: newMsgHeader(MsgMPNews, tagID, userIds),
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
		msgPreviewHeader: newMsgPreviewHeader(MsgMPNews, wxName),
		News: news{
			MediaId: mediaId,
		},
	}

	id, dataId, err = c.send(&msg, SendForPreview)
	return
}

func (c *Client) SendNewsAll(mediaId string, tagID ...int) (id, dataId int, err error) {
	return c.sendNews(mediaId, tagID, nil)
}

func (c *Client) SendNewsByUsers(mediaId string, userIds []string) (id, dataId int, err error) {
	return c.sendNews(mediaId, nil, userIds)
}

type CorpArticle struct {
	Title string `json:"title"`
	Description string `json:"description"`
	URL string `json:"url"`
	PicURL string `json:"picurl"`
	BtnTxt string `json:"btntxt"`
}

type CorpNews struct {
	Articles []CorpArticle `json:"articles"`
}

func (c *Client) corpSendNews(news CorpNews, userIDs, partyIDs, tagIDs []string) error {
	var msg = struct{
		*corpMsgHeader
		CorpNews CorpNews `json:"news"`
	}{
		corpMsgHeader: c.newCorpMsgHeader(MsgNews, userIDs, partyIDs, tagIDs),
		CorpNews: news,
	}
	return c.corpSend(&msg)
}

func (c *Client) CorpSendNewsToAll(news CorpNews) error {
	return c.corpSendNews(news, []string{"@all"}, nil, nil)
}

func (c *Client) CorpSendNewsByUsers(news CorpNews, userIDs []string) error {
	return c.corpSendNews(news, userIDs, nil, nil)
}

func (c *Client) CorpSendNewsByParties(news CorpNews, partyIDs []string) error {
	return c.corpSendNews(news, nil, partyIDs, nil)
}

func (c *Client) CorpSendNewsByTags(news CorpNews, tagIDs []string) error {
	return c.corpSendNews(news, nil, nil, tagIDs)
}

type CorpMPArticle struct {
	Title            string `json:"title"`
	Author           string `json:"author,omitempty"`
	ThumbId          string `json:"thumb_media_id"` // media id of cover picture
	ContentSourceURL string `json:"content_source_url,omitempty"` // URL of "Read the original content"
	Content          string `json:"content"`                      // content, supporting HTML, no JS, less than 666KB
	Digest           string `json:"digest,omitempty"` // content digest, less than 512B
}

type CorpMPNews struct {
	Articles []CorpMPArticle `json:"articles"`
}

func (c *Client) corpSendMPNews(news CorpMPNews, userIDs, partyIDs, tagIDs []string) error {
	var msg = struct{
		*corpMsgHeader
		CorpMPNews CorpMPNews `json:"mpnews"`
	}{
		corpMsgHeader: c.newCorpMsgHeader(MsgMPNews, userIDs, partyIDs, tagIDs),
		CorpMPNews: news,
	}
	return c.corpSend(&msg)
}

func (c *Client) CorpSendMPNewsToAll(news CorpMPNews) error {
	return c.corpSendMPNews(news, []string{"@all"}, nil, nil)
}

func (c *Client) CorpSendMPNewsByUsers(news CorpMPNews, userIDs []string) error {
	return c.corpSendMPNews(news, userIDs, nil, nil)
}

func (c *Client) CorpSendMPNewsByParties(news CorpMPNews, partyIDs []string) error {
	return c.corpSendMPNews(news, nil, partyIDs, nil)
}

func (c *Client) CorpSendMPNewsByTags(news CorpMPNews, tagIDs []string) error {
	return c.corpSendMPNews(news, nil, nil, tagIDs)
}

type Card struct {
	Id  string `json:"card_id"`
	Ext string `json:"card_ext,omitempty"`
}

func (c *Client) sendCard(cardId, cardExt string, tagID []int, userIds []string) (id int, err error) {
	var msg = struct {
		*msgHeader
		Card Card `json:"wxcard"`
	}{
		msgHeader: newMsgHeader(MsgCard, tagID, userIds),
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

func (c *Client) SendCardAll(cardId, cardExt string, tagID ...int) (id int, err error) {
	return c.sendCard(cardId, cardExt, tagID, nil)
}

func (c *Client) SendCardByUsers(cardId, cardExt string, userIds []string) (id int, err error) {
	return c.sendCard(cardId, cardExt, nil, userIds)
}

type TextCard struct {
	Title string `json:"title"`
	Description string `json:"description"`
	URL string `json:"url"`
	BtnTxt string `json:"btntxt"`
}

func (c *Client) corpSendCard(card TextCard, userIDs, partyIDs, tagIDs []string) error {
	var msg = struct{
		*corpMsgHeader
		TextCard TextCard `json:"textcard"`
	}{
		corpMsgHeader: c.newCorpMsgHeader(MsgMPNews, userIDs, partyIDs, tagIDs),
		TextCard: card,
	}
	return c.corpSend(&msg)
}

func (c *Client) CorpSendCardToAll(card TextCard) error {
	return c.corpSendCard(card, []string{"@all"}, nil, nil)
}

func (c *Client) CorpSendCardByUsers(card TextCard, userIDs []string) error {
	return c.corpSendCard(card, userIDs, nil, nil)
}

func (c *Client) CorpSendCardByParties(card TextCard, partyIDs []string) error {
	return c.corpSendCard(card, nil, partyIDs, nil)
}

func (c *Client) CorpSendCardByTags(card TextCard, tagIDs []string) error {
	return c.corpSendCard(card, nil, nil, tagIDs)
}

// DeleteMsg deletes the mass message(only MsgMPNews and MsgMPVideo) which was sent in half an hour.
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
