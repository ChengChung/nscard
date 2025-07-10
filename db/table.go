package db

import "time"

type GameTitleImgCache struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	TitleName string `gorm:"not null"`
	ImgUrl    string `gorm:"not null"`
	Data      string `gorm:"not null"`
}

func (GameTitleImgCache) TableName() string {
	return "game_title_img_cache"
}

type SwitchCardUser struct {
	ID                int64  `gorm:"primaryKey;autoIncrement"`
	UserId            string `gorm:"not null"`
	SessionToken      string `gorm:"not null"`
	Nickname          string `gorm:"not null"`
	Country           string `gorm:"not null"`
	NintendoAvatarUrl string `gorm:"not null"`
	CustomAvatarUrl   string
	SwFriendCode      string
	CustomGame        string
	CustomBackground  string
	CreatedTime       time.Time `gorm:"default:current_timestamp"`
	UpdatedTime       time.Time `gorm:"default:current_timestamp"`
}

func (SwitchCardUser) TableName() string {
	return "switch_card_user"
}

type JPStoreGameTitle struct {
	ID               int64  `gorm:"primaryKey;autoIncrement"`
	InitialCode      string `gorm:"column:initial_code"`
	TitleName        string `gorm:"column:title_name"`
	MakerName        string `gorm:"column:maker_name"`
	MakerKana        string `gorm:"column:maker_kana"`
	Price            string `gorm:"column:price"`
	SalesDate        string `gorm:"column:sales_date"`
	SoftType         string `gorm:"column:soft_type"`
	PlatformID       string `gorm:"column:platform_id"`
	DlIconFlg        string `gorm:"column:dl_icon_flg"`
	LinkURL          string `gorm:"column:link_url"`
	ScreenshotImgFlg string `gorm:"column:screenshot_img_flg"`
	ScreenshotImgURL string `gorm:"column:screenshot_img_url"`
}

func (JPStoreGameTitle) TableName() string {
	return "jp_store_game_title"
}

type JPStoreGameTitleSearch struct {
	ID     string `gorm:"primaryKey;not null"`
	Title  string `gorm:"column:title"`
	URL    string `gorm:"column:url"`
	TitleK string `gorm:"column:titlek"`
	NSUID  string `gorm:"column:nsuid"`
	Hard   string `gorm:"column:hard"`
	IURL   string `gorm:"column:iurl"`
	SIURL  string `gorm:"column:siurl"`
}

func (JPStoreGameTitleSearch) TableName() string {
	return "jp_store_game_title_search"
}

type SwitchCardCache struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	UserId      string    `gorm:"not null"`
	Cache       string    `gorm:"not null"`
	UpdatedTime time.Time `gorm:"not null"`
}

func (SwitchCardCache) TableName() string {
	return "switch_card_cache"
}
