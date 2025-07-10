package card

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chengchung/nscard/cache"
	"github.com/chengchung/nscard/client/nintendo/auth"
	"github.com/chengchung/nscard/client/nintendo/user"
	"github.com/chengchung/nscard/db"
	"github.com/chengchung/nscard/utils"
	"github.com/sirupsen/logrus"
)

type CardInfo struct {
	AvatarData     string
	FlagData       string
	BackgroundData string

	NickName     string
	SwFriendCode string

	HistoryData
}

func RenderCard(uid string) error {
	cli, ok := cache.GetAuthClient(uid)
	if !ok {
		return errors.New("user not found")
	}

	userInfo, ok := cache.GetUserInfo(uid)
	if !ok {
		return errors.New("user info not found")
	}

	rawHistory, err := auth.WithAccessToken(user.GetPlayHistory)(cli)
	if err != nil {
		logrus.Errorf("failed to get play history for user %s: %v", uid, err)
		return err
	}
	history, err := collectHistory(rawHistory, userInfo.CustomGame, 6)
	if err != nil {
		logrus.Errorf("failed to collect history for user %s: %v", uid, err)
		return err
	}

	images, err := collectImages(userInfo, history)
	if err != nil {
		logrus.Errorf("failed to collect images for user %s: %v", uid, err)
		return err
	}

	for idx := range history.FirstNGames {
		game := &history.FirstNGames[idx]
		if iconData, ok := images.games_icon_data[game.TitleName]; ok {
			game.ImageData = iconData
		}
	}
	cardInfo := CardInfo{
		AvatarData:     images.avatar_data,
		NickName:       userInfo.Nickname,
		FlagData:       images.flag_data,
		SwFriendCode:   userInfo.SwFriendCode,
		BackgroundData: images.card_bg_data,
		HistoryData:    *history,
	}

	svg, err := RenderSVG(cardInfo)
	if err != nil {
		logrus.Errorf("failed to render SVG for user %s: %v", uid, err)
		return err
	}

	if err := db.SetSwitchCardCache(uid, svg); err != nil {
		logrus.Errorf("failed to set switch card cache for user %s: %v", uid, err)
		return err
	}

	return nil
}

type HistoryData struct {
	GamesTotal     int
	PlayTimeTotal  string
	GamesThisMonth int
	GamesThisYear  int

	FirstNGames []GameBriefInfo
	CustomGame  *GameBriefInfo
}

type GameBriefInfo struct {
	TitleName string
	ImageURL  string
	ImageData string
	PlayTime  string

	fake bool
}

func collectHistory(history *user.UserPlayHistory, customGameName string, firstCnt int) (*HistoryData, error) {
	data := &HistoryData{
		GamesTotal: len(history.PlayHistories),
	}
	playTimeTotal := 0
	for idx, game := range history.PlayHistories {
		playTimeTotal += game.TotalPlayedMins
		lastPlayAt, err := time.Parse(time.RFC3339, game.LastPlayedAt)
		if err != nil {
			logrus.Warnf("failed to parse last played at for game %s: %v", game.TitleName, err)
			continue
		}
		now := time.Now()
		if lastPlayAt.Year() == now.Year() {
			data.GamesThisYear++
			if lastPlayAt.Month() == now.Month() {
				data.GamesThisMonth++
			}
		}

		if idx < firstCnt {
			data.FirstNGames = append(data.FirstNGames, GameBriefInfo{
				TitleName: game.TitleName,
				ImageURL:  game.ImageURL,
				PlayTime:  formatMin(game.TotalPlayedMins),
			})
		}
		if game.TitleName == customGameName {
			data.CustomGame = &GameBriefInfo{
				TitleName: game.TitleName,
				ImageURL:  game.ImageURL,
				PlayTime:  formatMin(game.TotalPlayedMins),
			}
		}
	}
	data.PlayTimeTotal = formatMin(playTimeTotal)

	return data, nil
}

type ImageInfo struct {
	avatar_data     string
	flag_data       string
	card_bg_data    string
	games_icon_data map[string]string
}

func collectImages(userInfo *cache.CacheUserInfo, history *HistoryData) (*ImageInfo, error) {
	avatar_url := userInfo.NintendoAvatarUrl
	if len(userInfo.CustomAvatarUrl) > 0 {
		avatar_url = userInfo.CustomAvatarUrl
	}
	cflag_url := getCFlagSVG(userInfo.Country)

	requests := make([]cache.ImageRequest, 0, len(history.FirstNGames)+3)
	requests = append(requests, cache.ImageRequest{
		TitleName: "__internal_app_avatar__",
		ImageUrl:  avatar_url,
	}, cache.ImageRequest{
		TitleName: "__internal_app_cflag__",
		ImageUrl:  cflag_url,
	})

	var card_bg_url string

	bg_games := make([]GameBriefInfo, 0, 3)
	if history.CustomGame != nil {
		bg_games = append(bg_games, *history.CustomGame)
	}
	if len(userInfo.CustomBackground) > 0 {
		bg_games = append(bg_games, GameBriefInfo{
			TitleName: "__user_custom_background__",
			ImageURL:  userInfo.CustomBackground,
			fake:      true,
		})
	}
	if len(history.FirstNGames) > 0 {
		bg_games = append(bg_games, history.FirstNGames[0])
	}

	for _, game := range bg_games {
		card_bg_url = getCardBGUrl(game)
		if len(card_bg_url) > 0 {
			requests = append(requests, cache.ImageRequest{
				TitleName: game.TitleName,
				ImageUrl:  card_bg_url,
			})
			break
		}
	}

	for _, game := range history.FirstNGames {
		requests = append(requests, cache.ImageRequest{
			TitleName: game.TitleName,
			ImageUrl:  game.ImageURL,
		})
	}

	resps, err := cache.GetOrCreateImageCache(requests)
	if err != nil {
		logrus.Errorf("failed to collect images for user %s: %v", userInfo.UserId, err)
		return nil, err
	}
	logrus.Infof("collected %d images for user %s", len(resps), userInfo.UserId)

	imagesURL_Data := make(map[string][]byte, len(resps))
	for _, resp := range resps {
		imagesURL_Data[resp.ImageUrl] = resp.Data
	}

	info := ImageInfo{
		flag_data:       utils.EncodeToBase64(utils.ImageTypeSVG, imagesURL_Data[cflag_url]),
		card_bg_data:    resizeAndIgnoreError(imagesURL_Data[card_bg_url], CARD_WIDTH, CARD_HEIGHT),
		avatar_data:     resizeAndIgnoreError(imagesURL_Data[avatar_url], AVATAR_WIDTH, AVATAR_HEIGHT),
		games_icon_data: make(map[string]string, len(history.FirstNGames)),
	}
	for _, game := range history.FirstNGames {
		if data, ok := imagesURL_Data[game.ImageURL]; ok {
			info.games_icon_data[game.TitleName] = resizeAndIgnoreError(data, ICON_WIDTH, ICON_HEIGHT)
		}
	}

	return &info, nil
}

func resizeAndIgnoreError(imgData []byte, width, height int) string {
	if len(imgData) == 0 {
		return ""
	}
	resizedData, err := utils.ResizeImage(bytes.NewReader(imgData), width, height)
	if err != nil {
		logrus.Errorf("failed to resize image: %v", err)
		return ""
	}
	return utils.EncodeToBase64(utils.ImageTypePNG, resizedData)
}

func getCardBGUrl(game GameBriefInfo) string {
	if game.fake {
		return game.ImageURL
	}

	title, _ := db.GetGameTitle(game.TitleName)
	if title != nil && len(title.ScreenshotImgURL) > 0 {
		return title.ScreenshotImgURL
	}
	titleSearch, _ := db.GetGameTitleSearch(game.TitleName)
	if titleSearch != nil && len(titleSearch.IURL) > 0 {
		return handleSearchImgUrl(titleSearch.IURL)
	}

	return getUnresizedIconUrl(game.ImageURL)
}

func getCFlagSVG(c string) string {
	return fmt.Sprintf(`https://cdn.jsdelivr.net/npm/round-flag-icons@1.3.0/flags/%s.svg`, strings.ToLower(c))
}

func getUnresizedIconUrl(imgUrl string) string {
	if strings.HasSuffix(imgUrl, "_256") {
		return strings.TrimSuffix(imgUrl, "_256")
	}
	return imgUrl
}

func handleSearchImgUrl(iurl string) string {
	if strings.HasPrefix(iurl, "/") {
		return "https://www.nintendo.com/jp" + iurl
	} else {
		return "https://img-eshop.cdn.nintendo.net/i/" + iurl + ".jpg"
	}
}

func formatMin(min int) string {
	if min < 60 {
		return fmt.Sprintf("%dmin", min)
	} else {
		return fmt.Sprintf("%.1fh", float64(min)/60)
	}
}
