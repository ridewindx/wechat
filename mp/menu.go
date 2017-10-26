package mp

const (
	// 下面6个类型(包括view类型)的按钮是在公众平台官网发布的菜单按钮类型
	ButtonTypeText  = "text"
	ButtonTypeImage = "img"
	ButtonTypePhoto = "photo"
	ButtonTypeVideo = "video"
	ButtonTypeVoice = "voice"
	// 上面5个类型的按钮不能通过API设置

	ButtonTypeView  = "view"  // 跳转URL
	ButtonTypeClick = "click" // 点击推事件

	ButtonTypeMiniprogram = "miniprogram" // 小程序

	// 下面的按钮类型仅支持微信 iPhone5.4.1 以上版本, 和 Android5.4 以上版本的微信用户,
	// 旧版本微信用户点击后将没有回应, 开发者也不能正常接收到事件推送
	ButtonTypeScanCodePush    = "scancode_push"      // 扫码推事件
	ButtonTypeScanCodeWaitMsg = "scancode_waitmsg"   // 扫码带提示
	ButtonTypePicSysPhoto     = "pic_sysphoto"       // 系统拍照发图
	ButtonTypePicPhotoOrAlbum = "pic_photo_or_album" // 拍照或者相册发图
	ButtonTypePicWeixin       = "pic_weixin"         // 微信相册发图
	ButtonTypeLocationSelect  = "location_select"    // 发送位置

	// 下面的按钮类型专门给第三方平台旗下未微信认证(具体而言, 是资质认证未通过)的订阅号准备的事件类型,
	// 它们是没有事件推送的, 能力相对受限, 其他类型的公众号不必使用
	ButtonTypeMediaId     = "media_id"     // 下发消息
	ButtonTypeViewLimited = "view_limited" // 跳转图文消息URL
)

type Menu struct {
	Buttons   []Button   `json:"button,omitempty"`
	MatchRule *MatchRule `json:"matchrule,omitempty"`
	MenuId    int64      `json:"menuid,omitempty"`
}

type MatchRule struct {
	GroupID            *int64 `json:"group_id,omitempty"`
	Sex                *int   `json:"sex,omitempty"`
	Country            string `json:"country,omitempty"`
	Province           string `json:"province,omitempty"`
	City               string `json:"city,omitempty"`
	ClientPlatformType *int   `json:"client_platform_type,omitempty"`
	Language           string `json:"language,omitempty"`
}

type Button struct {
	Type       string   `json:"type,omitempty"`
	SubButtons []Button `json:"sub_button,omitempty"`
	Name       string   `json:"name,omitempty"`     // 菜单标题，不超过16个字节，子菜单不超过60个字节
	Key        string   `json:"key,omitempty"`      // 菜单KEY值，用于消息接口推送，不超过128字节
	URL        string   `json:"url,omitempty"`      // 网页链接，用户点击菜单可打开链接，不超过1024字节
	MediaID    string   `json:"media_id,omitempty"` // 调用新增永久素材接口返回的合法media_id
	AppID      string   `json:"appid,omitempty"`    // 小程序AppID（仅认证公众号可配置）
	PagePath   string   `json:"pagepath"`           // 小程序的页面路径
}

func NewButtonWithSubButtons(name string, subButtons []Button) *Button {
	return &Button{
		Name:       name,
		SubButtons: subButtons,
	}
}

func NewButtonForClick(name, key string) *Button {
	return &Button{
		Type: ButtonTypeClick,
		Name: name,
		Key:  key,
	}
}

func NewButtonForView(name, url string) *Button {
	return &Button{
		Type: ButtonTypeView,
		Name: name,
		URL:  url,
	}
}

func NewButtonForMiniprogram(name, url, appID, pagePath string) *Button {
	return &Button{
		Type:     ButtonTypeMiniprogram,
		Name:     name,
		URL:      url,
		AppID:    appID,
		PagePath: pagePath,
	}
}

func NewButtonForScanCodePush(name, key string) *Button {
	return &Button{
		Type: ButtonTypeScanCodePush,
		Name: name,
		Key:  key,
	}
}

func NewButtonForScanCodeWaitMsg(name, key string) *Button {
	return &Button{
		Type: ButtonTypeScanCodeWaitMsg,
		Name: name,
		Key:  key,
	}
}

func NewButtonForPicSysPhoto(name, key string) *Button {
	return &Button{
		Type: ButtonTypePicSysPhoto,
		Name: name,
		Key:  key,
	}
}

func NewButtonForPicPhotoOrAlbum(name, key string) *Button {
	return &Button{
		Type: ButtonTypePicPhotoOrAlbum,
		Name: name,
		Key:  key,
	}
}

func NewButtonForPicWeixin(name, key string) *Button {
	return &Button{
		Type: ButtonTypePicWeixin,
		Name: name,
		Key:  key,
	}
}

func NewButtonForLocationSelect(name, key string) *Button {
	return &Button{
		Type: ButtonTypeLocationSelect,
		Name: name,
		Key:  key,
	}
}

func NewButtonForMediaId(name, mediaID string) *Button {
	return &Button{
		Type:    ButtonTypeMediaId,
		Name:    name,
		MediaID: mediaID,
	}
}

func NewButtonForViewLimited(name, mediaID string) *Button {
	return &Button{
		Type:    ButtonTypeViewLimited,
		Name:    name,
		MediaID: mediaID,
	}
}

func (c *Client) CreateMenu(menu *Menu) error {
	u := BASE_URL.Join("/menu/create")

	var rep Err
	return c.Post(u, menu, &rep)
}

func (c *Client) CreateConditionalMenu(menu *Menu) (menuID int64, err error) {
	u := BASE_URL.Join("/menu/addconditional")

	var rep struct {
		Err
		MenuID int64 `json:"menuId"`
	}

	err = c.Post(u, menu, &rep)
	if err != nil {
		return
	}
	menuID = rep.MenuID
	return
}

func (c *Client) GetMenus() (menu *Menu, conditionalMenus []Menu, err error) {
	u := BASE_URL.Join("/menu/get")

	var rep struct {
		Err
		Menu             Menu   `json:"menu"`
		ConditionalMenus []Menu `json:"conditionalmenu"`
	}

	err = c.Get(u, &rep)
	if err != nil {
		return
	}

	return &rep.Menu, rep.ConditionalMenus, nil
}

func (c *Client) DeleteMenu() error {
	u := BASE_URL.Join("/menu/delete")

	var rep Err
	return c.Get(u, &rep)
}

func (c *Client) DeleteConditionalMenu(menuID *Menu) error {
	u := BASE_URL.Join("/menu/delconditional")

	var req struct{
		MenuID int64 `json:"menuId"`
	}

	var rep Err

	err := c.Post(u, &req, &rep)
	return err
}
