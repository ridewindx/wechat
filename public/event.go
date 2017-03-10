package public

// Received message types
const (
	MessageText  = "text"
	MessageImage = "image"
	MessageVoice = "voice"
	MessageVideo = "video"
	MessageMusic = "music"
	MessageNews  = "news"
	MessageEvent = "event"
)

// Event types for MessageEvent
const (
	EventSubscribe                   = "subscribe"
	EventUnsubscribe                 = "unsubscribe"
	EventScan                        = "SCAN"
	EventLocation                    = "LOCATION"
	EventMenuClick                   = "CLICK"
	EventMenuView                    = "VIEW"
	EventCreateAgentSession          = "kf_create_session"
	EventCloseAgentSession           = "kf_close_session"
	EventSwitchAgentSession          = "kf_switch_session"
	EventQualificationVerifySuccess  = "qualification_verify_success"
	EventQualificationVerifyFail     = "qualification_verify_fail"
	EventNamingVerifySuccess         = "naming_verify_success"
	EventNamingVerifyFail            = "naming_verify_fail"
	EventAnnualRenew                 = "annual_renew"
	EventVerifyExpired               = "verify_expired"
	EventCardPassCheck               = "card_pass_check"
	EventNotCardPassCheck            = "card_not_pass_check"
	EventUserGetCard                 = "user_get_card"
	EventUserDelCard                 = "user_del_card"
	EventUserConsumeCard             = "user_consume_card"
	EventUserPayFromPayCell          = "user_pay_from_pay_cell"
	EventUserViewCard                = "user_view_card"
	EventUserEnterSessionFromCard    = "user_enter_session_from_card"
	EventUpdateMemberCard            = "update_member_card"
	EventCardSkuRemind               = "card_sku_remind"
	EventCardPayOrder                = "card_pay_order"
	EventUserScanProduct             = "user_scan_product"
	EventUserScanProductEnterSession = "user_scan_product_enter_session"
	EventUserScanProductAsync        = "user_scan_product_async"
	EventUserScanProductVerifyAction = "user_scan_product_verify_action"
	EventShakeAroundUserShake        = "ShakearoundUserShake"
)

type EventHeader struct {
	ToUser      string `xml:"ToUserName"   json:"ToUserName"`
	FromUser    string `xml:"FromUserName" json:"FromUserName"`
	CreatedTime int64  `xml:"CreateTime"   json:"CreateTime"`
	Type        string `xml:"MsgType"      json:"MsgType"`
}

type Event struct {
	EventHeader

	Event string `xml:"Event" json:"Event"`

	MsgId        int     `xml:"MsgId"        json:"MsgId"`
	Content      string  `xml:"Content"      json:"Content"`
	MediaId      string  `xml:"MediaId"      json:"MediaId"`
	PicURL       string  `xml:"PicUrl"       json:"PicUrl"`
	Format       string  `xml:"Format"       json:"Format"`
	Recognition  string  `xml:"Recognition"  json:"Recognition"`
	ThumbMediaId string  `xml:"ThumbMediaId" json:"ThumbMediaId"`
	LocationX    float64 `xml:"Location_X"   json:"Location_X"`
	LocationY    float64 `xml:"Location_Y"   json:"Location_Y"`
	Scale        int     `xml:"Scale"        json:"Scale"`
	Label        string  `xml:"Label"        json:"Label"`
	Title        string  `xml:"Title"        json:"Title"`
	Description  string  `xml:"Description"  json:"Description"`
	URL          string  `xml:"Url"          json:"Url"`
	EventKey     string  `xml:"EventKey"     json:"EventKey"`
	Ticket       string  `xml:"Ticket"       json:"Ticket"`
	Latitude     float64 `xml:"Latitude"     json:"Latitude"`
	Longitude    float64 `xml:"Longitude"    json:"Longitude"`
	Precision    float64 `xml:"Precision"    json:"Precision"`

	MenuId       int `xml:"MenuId" json:"MenuId"`
	ScanCodeInfo *struct {
		ScanType   string `xml:"ScanType"   json:"ScanType"`
		ScanResult string `xml:"ScanResult" json:"ScanResult"`
	} `xml:"ScanCodeInfo,omitempty" json:"ScanCodeInfo,omitempty"`
	SendPicsInfo *struct {
		Count   int `xml:"Count" json:"Count"`
		PicList []struct {
			PicMd5Sum string `xml:"PicMd5Sum" json:"PicMd5Sum"`
		} `xml:"PicList>item,omitempty" json:"PicList,omitempty"`
	} `xml:"SendPicsInfo,omitempty" json:"SendPicsInfo,omitempty"`
	SendLocationInfo *struct {
		LocationX float64 `xml:"Location_X" json:"Location_X"`
		LocationY float64 `xml:"Location_Y" json:"Location_Y"`
		Scale     int     `xml:"Scale"      json:"Scale"`
		Label     string  `xml:"Label"      json:"Label"`
		PoiName   string  `xml:"Poiname"    json:"Poiname"`
	} `xml:"SendLocationInfo,omitempty" json:"SendLocationInfo,omitempty"`

	Status string `xml:"Status" json:"Status"`

	ChosenBeacon  *Beacon `xml:"ChosenBeacon,omitempty" json:"ChosenBeacon,omitempty"`
	AroundBeacons *Beacon `xml:"AroundBeacons>AroundBeacon,omitempty" json:"AroundBeacons,omitempty"`
}

type Beacon struct {
	UUID     string  `xml:"Uuid"     json:"Uuid"`
	Major    int     `xml:"Major"    json:"Major"`
	Minor    int     `xml:"Minor"    json:"Minor"`
	Distance float64 `xml:"Distance" json:"Distance"`
}

func responseEventHeader(msgType string, event *Event) *EventHeader {
	return &EventHeader{
		ToUser:      event.FromUser,
		FromUser:    event.ToUser,
		CreatedTime: event.CreatedTime,
		Type:        msgType,
	}
}
