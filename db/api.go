package db

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetAllSwitchCardUsers() ([]SwitchCardUser, error) {
	var users []SwitchCardUser
	err := GetConn().Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func CreateOrUpdateSwitchCardUser(user *SwitchCardUser) error {
	// update if UserId exists, otherwise create
	err := GetConn().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"session_token", "nickname", "country", "nintendo_avatar_url", "updated_time"}),
	}, clause.Returning{}).Create(user).Error

	return err
}

func UpdateUserSettings(userId string, user *SwitchCardUser) error {
	err := GetConn().Model(user).
		Where("user_id = ?", userId).
		Select("custom_avatar_url", "sw_friend_code", "custom_game", "custom_background", "updated_time").
		Clauses(clause.Returning{}).
		Updates(user).Error

	return err
}

func GetSwitchCardUsers(userId string) (*SwitchCardUser, error) {
	var user SwitchCardUser
	err := GetConn().Where(&SwitchCardUser{UserId: userId}).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func BatchSaveJPStoreGameTitle(titles []JPStoreGameTitle) error {
	err := GetConn().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "initial_code"}, {Name: "title_name"}, {Name: "soft_type"}, {Name: "platform_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"maker_name", "maker_kana", "price", "sales_date", "dl_icon_flg", "link_url", "screenshot_img_flg", "screenshot_img_url"}),
	}).CreateInBatches(&titles, 65535/13).Error
	if err != nil {
		return err
	}

	return nil
}

func BatchSaveJPStoreGameTitleSearch(titles []JPStoreGameTitleSearch) error {
	err := GetConn().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "url", "titlek", "nsuid", "hard", "iurl", "siurl"}),
	}).CreateInBatches(&titles, 65535/8).Error
	if err != nil {
		return err
	}

	return nil
}

func GetSwitchCardCache(userId string) (*SwitchCardCache, error) {
	var cache SwitchCardCache
	err := GetConn().Where(&SwitchCardCache{UserId: userId}).First(&cache).Error
	if err != nil {
		return nil, err
	}
	return &cache, nil
}

func SetSwitchCardCache(userid string, img string) error {
	err := GetConn().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"cache", "updated_time"}),
	}).Create(&SwitchCardCache{
		UserId:      userid,
		Cache:       img,
		UpdatedTime: time.Now(),
	}).Error
	return err
}

func GetGameTitleImgCache(urls []string) ([]GameTitleImgCache, error) {
	var caches []GameTitleImgCache
	err := GetConn().Where("img_url IN ?", urls).Find(&caches).Error
	if err != nil {
		return nil, err
	}
	return caches, nil
}

func BatchSetGameTitleImageCache(caches []GameTitleImgCache) error {
	err := GetConn().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "title_name"}, {Name: "img_url"}},
		DoUpdates: clause.AssignmentColumns([]string{"data"}),
	}).CreateInBatches(&caches, 65535/4).Error
	if err != nil {
		return err
	}

	return nil
}

func GetGameTitle(titleName string) (*JPStoreGameTitle, error) {
	var title JPStoreGameTitle
	err := GetConn().Where(&JPStoreGameTitle{TitleName: titleName}).First(&title).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &title, nil
}

func GetGameTitleSearch(titleName string) (*JPStoreGameTitleSearch, error) {
	var titleSearch JPStoreGameTitleSearch
	err := GetConn().Where(&JPStoreGameTitleSearch{Title: titleName}).First(&titleSearch).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &titleSearch, nil
}
