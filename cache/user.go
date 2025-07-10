package cache

import (
	"sync"
	"time"

	"github.com/chengchung/nscard/client/nintendo/auth"
	"github.com/chengchung/nscard/client/nintendo/user"
	"github.com/chengchung/nscard/db"
	"github.com/sirupsen/logrus"
)

var user_map sync.Map

type cacheUser struct {
	cli      *auth.AccessClient
	userInfo *CacheUserInfo
}

type CacheUserInfo struct {
	UserId            string
	Nickname          string
	Country           string
	NintendoAvatarUrl string
	UserCustomSettings
}

type UserCustomSettings struct {
	CustomAvatarUrl  string `json:"custom_avatar_url"`
	SwFriendCode     string `json:"sw_friend_code"`
	CustomGame       string `json:"custom_game"`
	CustomBackground string `json:"custom_background"`
}

func CreateOrUpdateUserCache(userInfo *user.UserInfo, sessionToken string) error {
	now := time.Now()
	user := db.SwitchCardUser{
		UserId:            userInfo.ID,
		Nickname:          userInfo.Nickname,
		Country:           userInfo.Country,
		NintendoAvatarUrl: userInfo.IconUri,
		SessionToken:      sessionToken,
		CreatedTime:       now,
		UpdatedTime:       now,
	}
	if err := db.CreateOrUpdateSwitchCardUser(&user); err != nil {
		logrus.Errorf("failed to create or update user %s in db: %v", user.UserId, err)
		return err
	}

	restoreUserCache(&user)

	return nil
}

func UpdateUserSettings(userId string, userInfo *UserCustomSettings) error {
	now := time.Now()
	user := db.SwitchCardUser{
		UserId:           userId,
		CustomAvatarUrl:  userInfo.CustomAvatarUrl,
		SwFriendCode:     userInfo.SwFriendCode,
		CustomGame:       userInfo.CustomGame,
		CustomBackground: userInfo.CustomBackground,
		UpdatedTime:      now,
	}

	if err := db.UpdateUserSettings(userId, &user); err != nil {
		logrus.Errorf("failed to update user %s settings in db: %v", userId, err)
		return err
	}

	logrus.Infof("User new settings: %#v", user)

	restoreUserCache(&user)
	return nil
}

func restoreUserCache(user *db.SwitchCardUser) *cacheUser {
	n := cacheUser{
		cli: auth.NewAccessClient(user.SessionToken),
		userInfo: &CacheUserInfo{
			UserId:            user.UserId,
			Nickname:          user.Nickname,
			Country:           user.Country,
			NintendoAvatarUrl: user.NintendoAvatarUrl,
			UserCustomSettings: UserCustomSettings{
				CustomAvatarUrl:  user.CustomAvatarUrl,
				SwFriendCode:     user.SwFriendCode,
				CustomGame:       user.CustomGame,
				CustomBackground: user.CustomBackground,
			},
		},
	}
	user_map.Store(user.UserId, &n)

	return &n
}

func GetAuthClient(userId string) (*auth.AccessClient, bool) {
	actual, ok := user_map.Load(userId)
	if !ok {
		cacheUser, found := restoreUserCacheFromDB(userId)
		if !found {
			return nil, false
		}
		return cacheUser.cli, true
	}

	c := actual.(*cacheUser)
	if c.cli == nil {
		return nil, false
	}

	return c.cli, true
}

func GetUserInfo(userId string) (*CacheUserInfo, bool) {
	actual, ok := user_map.Load(userId)
	if !ok {
		cacheUser, found := restoreUserCacheFromDB(userId)
		if !found {
			return nil, false
		}
		return cacheUser.userInfo, true
	}

	c := actual.(*cacheUser)
	if c.userInfo == nil {
		return nil, false
	}

	return c.userInfo, true
}

func restoreUserCacheFromDB(userId string) (*cacheUser, bool) {
	userFound, err := db.GetSwitchCardUsers(userId)
	if err != nil || userFound == nil {
		logrus.Errorf("failed to get user %s from db: %v", userId, err)
		return nil, false
	}

	cache := restoreUserCache(userFound)
	return cache, true
}
