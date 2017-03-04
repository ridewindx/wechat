package public

import (
	"mime/multipart"
)

const (
	Image = "image"
	Voice = "voice"
	Video = "video"
	Thumb = "thumb"
	News  = "news"
)

type MediaInfo struct {
	MediaType string `json:"type"`
	MediaId   string `json:"media_id"`
	CreatedAt int64  `json:"created_at"`
}


