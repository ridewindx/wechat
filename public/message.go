package public

const (
	MsgText  = "text"
	MsgImage = "image"
	MsgVoice = "voice"
	MsgVideo = "mpvideo"
	MsgNews  = "mpnews"
	MsgCard  = "wxcard"
)

type MsgHeader struct {
	Filter struct {
		IsToAll bool `json:"is_to_all"`
		GroupId bool `json:"group_id"`
	} `json:"filter"`
	MsgType string `json:"msgtype"`
	ToUser  []string     `json:"touser,omitempty"`
}

type Text struct {
	MsgHeader
	Text struct {
		Content string `json:"content"`
	} `json:"text"`
}

type Image struct {
	MsgHeader
	Image struct {
		MediaId string `json:"media_id"`
	} `json:"image"`
}

type Voice struct {
	MsgHeader
	Voice struct {
		MediaId string `json:"media_id"`
	} `json:"voice"`
}

type Video struct {
	MsgHeader
	Video struct {
		MediaId string `json:"media_id"`
	} `json:"mpvideo"`
}

type News struct {
	MsgHeader
	News struct {
		MediaId string `json:"media_id"`
	} `json:"mpnews"`
}

type Card struct {
	MsgHeader
	Card struct {
		Id  string `json:"card_id"`
		Ext string `json:"card_ext,omitempty"`
	} `json:"wxcard"`
}
