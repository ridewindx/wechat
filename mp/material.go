package public

import (
	"encoding/json"
	"fmt"
	"github.com/kataras/go-errors"
)

const (
	MediaImage = "image"
	MediaVoice = "voice"
	MediaVideo = "video"
	MediaThumb = "thumb"
	MediaNews  = "news"
)

// Temporary media
type TempMedia struct {
	Type      string `json:"type"`
	Id        string `json:"media_id"`
	CreatedAt int64  `json:"created_at"`
}

// Permanent media
type Media struct {
	Id         string `json:"media_id"`
	URL        string `json:"url"`         // only nonempty when media type is Image or Thumb, and only available on *.qq.com
	Name       string `json:"name"`        // only nonempty when call GetMediaList
	UpdateTime int64  `json:"update_time"` // only nonempty when call GetMediaList
}

type Video struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"down_url"`
}

// Article of words and images
type Article struct {
	ThumbId          string `json:"thumb_media_id"` // permant media id of cover picture
	Title            string `json:"title"`
	Author           string `json:"author,omitempty"`
	Digest           string `json:"digest,omitempty"`             // only valid for single-article news
	Content          string `json:"content"`                      // content, supporting HTML, no JS, less than 20000 chars or 1MB
	ContentSourceURL string `json:"content_source_url,omitempty"` // URL of "Read the original content"
	ShowCoverPic     int    `json:"show_cover_pic"`               // whether show cover picture
	ThumbURL         string `json:"thumb_url"`                    // cover picture URL, only valid for GetNews
	URL              string `json:"url"`                          // content URL, only valid for GetNews
}

type News struct {
	Articles []Article `json:"articles,omitempty"`
}

type MediaCounts struct {
	VoiceCount int `json:"voice_count"`
	VideoCount int `json:"video_count"`
	ImageCount int `json:"image_count"`
	NewsCount  int `json:"news_count"`
}

type NewsList struct {
	TotalCount int `json:"total_count"`
	ItemCount  int `json:"item_count"` // item count of this time GetNewsList
	Items      []struct {
		Id         string `json:"media_id"`
		UpdateTime int64  `json:"update_time"`
		Content    struct {
			Articles []Article `json:"news_item,omitempty"`
		} `json:"content"`
	} `json:"item"`
}

type MediaList struct {
	TotalCount int     `json:"total_count"`
	ItemCount  int     `json:"item_count"` // item count of this time GetMediaList
	Items      []Media `json:"item"`
}

func (c *Client) UploadTempImage(filePath string) (*TempMedia, error) {
	return c.UploadTempMedia(MediaImage, filePath)
}

func (c *Client) UploadTempVoice(filePath string) (*TempMedia, error) {
	return c.UploadTempMedia(MediaVoice, filePath)
}

func (c *Client) UploadTempVideo(filePath string) (*TempMedia, error) {
	return c.UploadTempMedia(MediaVideo, filePath)
}

func (c *Client) UploadTempThumb(filePath string) (*TempMedia, error) {
	return c.UploadTempMedia(MediaThumb, filePath)
}

func (c *Client) UploadTempMedia(mediaType, filePath string) (*TempMedia, error) {
	u := BASE_URL.Join("/media/upload").Query("type", mediaType)

	var rep struct {
		Err
		TempMedia
	}

	err := c.UploadFile(u, "media", filePath, nil, &rep)
	if err != nil {
		return nil, err
	}

	return &rep.TempMedia, err
}

func (c *Client) DownloadTempMedia(mediaId, filePath string) error {
	u := BASE_URL.Join("/media/get").Query("media_id", mediaId) // TODO: download video needs http, not https

	var rep Err

	return c.DownloadFile(u, nil, filePath, &rep)
}

func (c *Client) UploadImage(filePath string) (*Media, error) {
	return c.UploadMedia(MediaImage, filePath)
}

func (c *Client) UploadThumb(filePath string) (*Media, error) {
	return c.UploadMedia(MediaThumb, filePath)
}

func (c *Client) UploadVoice(filePath string) (*Media, error) {
	return c.UploadMedia(MediaVoice, filePath)
}

func (c *Client) UploadVideo(title, intro, filePath string) (*Media, error) {
	var descr = struct {
		Title string `json:"title"`
		Intro string `json:"introduction"`
	}{
		Title: title,
		Intro: intro,
	}

	description, err := json.Marshal(&descr)
	if err != nil {
		return nil, err
	}

	extraFields := map[string]string{
		"description": string(description),
	}

	return c.UploadMedia(MediaVideo, filePath, extraFields)
}

func (c *Client) UploadMedia(mediaType, filePath string, extraFields ...map[string]string) (*Media, error) {
	u := BASE_URL.Join("/material/add_material").Query("type", mediaType)

	var rep struct {
		Err
		Media
	}

	var fields map[string]string
	if len(extraFields) > 0 {
		fields = extraFields[0]
	}

	err := c.UploadFile(u, "media", filePath, fields, &rep)
	if err != nil {
		return nil, err
	}

	return &rep.Media, err
}

func (c *Client) GetVideo(mediaId string) (video *Video, err error) {
	u := BASE_URL.Join("/material/get_material")

	var req = struct {
		Id string `json:"media_id"`
	}{
		Id: mediaId,
	}

	var rep struct {
		Err
		Video
	}

	err = c.Post(u, &req, &rep)
	if err != nil {
		return
	}

	video = &rep.Video
	return
}

func (c *Client) DownloadVideo(mediaId, filePath string) (err error) {
	video, err := c.GetVideo(mediaId)
	if err != nil {
		return
	}

	var rep Err

	err = c.DownloadFile(URL(video.URL), nil, filePath, &rep) // TODO: do not need token
	return
}

func (c *Client) DownloadMedia(mediaId, filePath string) (err error) {
	u := BASE_URL.Join("/material/get_material")

	var req = struct {
		Id string `json:"media_id"`
	}{
		Id: mediaId,
	}

	var rep Err

	err = c.DownloadFile(u, &req, filePath, &rep)
	return
}

func (c *Client) CreateNews(news *News) (mediaId string, err error) {
	u := BASE_URL.Join("/material/add_news")

	var rep struct {
		Err
		Id string `json:"media_id"`
	}

	err = c.Post(u, news, &rep)
	if err != nil {
		return
	}

	mediaId = rep.Id
	return
}

func (c *Client) GetNews(mediaId string) (news *News, err error) {
	u := BASE_URL.Join("/material/get_material")

	var req struct {
		Id string `json:"media_id"`
	}

	var rep struct {
		Err
		Articles []Article `json:"news_item"`
	}

	err = c.Post(u, &req, &rep)
	if err != nil {
		return
	}

	news = &News{
		Articles: rep.Articles,
	}
	return
}

// UpdateNews updates the index-th(0 based) article in the news which has media id mediaId.
func (c *Client) UpdateNews(mediaId string, index int, article *Article) (err error) {
	u := BASE_URL.Join("/material/update_news")

	var req = struct {
		Id      string   `json:"media_id"`
		Index   int      `json:"index"`
		Article *Article `json:"articles,omitempty"`
	}{
		Id:      mediaId,
		Index:   index,
		Article: article,
	}

	var rep Err

	err = c.Post(u, &req, &rep)
	return
}

func (c *Client) GetMediaCounts() (mediaCounts *MediaCounts, err error) {
	u := BASE_URL.Join("/material/get_materialcount")

	var rep struct {
		Err
		Counts MediaCounts
	}

	err = c.Get(u, &rep)
	if err != nil {
		return
	}

	mediaCounts = &rep.Counts
	return
}

func (c *Client) GetNewsList(offset, count int) (newsList *NewsList, err error) {
	u := BASE_URL.Join("/material/batchget_material")

	if count < 1 || count > 20 {
		err = errors.New("GetMediaList valid count range is [1,20]")
		return
	}

	var req = struct {
		Type   string `json:"type"`
		Offset int    `json:"offset"`
		Count  int    `json:"count"`
	}{
		Type:   MediaNews,
		Offset: offset,
		Count:  count,
	}

	var rep struct {
		Err
		NewsList
	}

	err = c.Post(u, &req, &rep)
	if err != nil {
		return
	}

	newsList = &rep.NewsList
	return
}

func (c *Client) GetMediaList(mediaType string, offset, count int) (mediaList *MediaList, err error) {
	u := BASE_URL.Join("/material/batchget_material")

	if mediaType == MediaVideo {
		err = fmt.Errorf("GetMediaList does not support mediaType '%s'", MediaVideo)
		return
	}

	if count < 1 || count > 20 {
		err = errors.New("GetMediaList valid count range is [1,20]")
		return
	}

	var req = struct {
		Type   string `json:"type"`
		Offset int    `json:"offset"`
		Count  int    `json:"count"`
	}{
		Type:   mediaType,
		Offset: offset,
		Count:  count,
	}

	var rep struct {
		Err
		MediaList
	}

	err = c.Post(u, &req, &rep)
	if err != nil {
		return
	}

	mediaList = &rep.MediaList
	return
}

func (c *Client) DeleteMedia(mediaId string) (err error) {
	u := BASE_URL.Join("/material/del_material")

	var req = struct {
		Id string `json:"media_id"`
	}{
		Id: mediaId,
	}

	var rep Err

	err = c.Post(u, &req, &rep)
	return
}
