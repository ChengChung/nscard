package user

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/chengchung/nscard/client/nintendo"
)

type UserPlayHistory struct {
	PlayHistories       []PlayHistoryRecord `json:"playHistories"`
	HiddenTitleList     []json.RawMessage   `json:"hiddenTitleList"` //	unknown
	RecentPlayHistories []RecentPlayHistoryRecord
}

type PlayHistoryRecord struct {
	TitleID         string `json:"titleId"`
	TitleName       string `json:"titleName"`
	Platform        string `json:"platform"`
	ImageURL        string `json:"imageUrl"`
	LastUpdatedAt   string `json:"lastUpdatedAt"`
	FirstPlayedAt   string `json:"firstPlayedAt"`
	LastPlayedAt    string `json:"lastPlayedAt"`
	TotalPlayedDays int    `json:"totalPlayedDays"`
	TotalPlayedMins int    `json:"totalPlayedMinutes"`
}

type RecentPlayHistoryRecord struct {
	PlayedDate         string                   `json:"playedDate"`
	DailyPlayHistories []DailyPlayHistoryRecord `json:"dailyPlayHistories"`
}

type DailyPlayHistoryRecord struct {
	TitleID         string `json:"titleId"`
	TitleName       string `json:"titleName"`
	Platform        string `json:"platform"`
	ImageURL        string `json:"imageUrl"`
	TotalPlayedMins int    `json:"totalPlayedMinutes"`
}

func GetPlayHistory(auth_header string) (*UserPlayHistory, error) {
	request, err := http.NewRequest("GET", "https://app-api.znej.nintendo.com/api/v2.0/users/me/play_histories", nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", auth_header)
	request.Header.Set("User-Agent", nintendo.UserAgent)
	request.Header.Set("gentry-locale", nintendo.GentryLocale)

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 >= 4 {
		return nil, errors.New("error status code: " + resp.Status + ", response text: " + string(bytes))
	}

	return ParsePlayHistory(bytes)
}

func ParsePlayHistory(data []byte) (*UserPlayHistory, error) {
	playHistory := UserPlayHistory{}
	err := json.Unmarshal(data, &playHistory)
	if err != nil {
		return nil, err
	}
	return &playHistory, nil
}
