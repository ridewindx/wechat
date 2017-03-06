package public

const (
	Image = "image"
	Voice = "voice"
	Video = "video"
	Thumb = "thumb"
	News  = "news"
)

type Media struct {
	Type      string `json:"type"`
	Id        string `json:"media_id"`
	CreatedAt int64  `json:"created_at"`
}

func (c *client) UploadImage(filePath string) (*Media, error) {
	return c.UploadMedia(Image, filePath)
}

func (c *client) UploadVoice(filePath string) (*Media, error) {
	return c.UploadMedia(Voice, filePath)
}

func (c *client) UploadVideo(filePath string) (*Media, error) {
	return c.UploadMedia(Video, filePath)
}

func (c *client) UploadThumb(filePath string) (*Media, error) {
	return c.UploadMedia(Thumb, filePath)
}

func (c *client) UploadMedia(mediaType, filePath string) (*Media, error) {
	u := BASE_URL.Join("/media/upload").Query("type", mediaType)

	var rep struct {
		Err
		Media
	}

	err := c.UploadFile(u, "media", filePath, &rep)
	if err != nil {
		return nil, err
	}

	return &rep.Media, err
}

func (c *client) DownloadMedia(mediaId, filePath string) error {
	u := BASE_URL.Join("/media/get").Query("media_id", mediaId) // TODO: download video needs http, not https

	var rep Err

	return c.DownloadFile(u, filePath, &rep)
}
