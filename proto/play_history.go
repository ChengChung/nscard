package proto

import "encoding/json"

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

type UserPlayHistory struct {
	PlayHistories       []PlayHistoryRecord `json:"playHistories"`
	HiddenTitleList     []json.RawMessage   `json:"hiddenTitleList"` //	unknown
	RecentPlayHistories []RecentPlayHistoryRecord
}

func ParsePlayHistory(data []byte) (*UserPlayHistory, error) {
	playHistory := UserPlayHistory{}
	err := json.Unmarshal(data, &playHistory)
	if err != nil {
		return nil, err
	}
	return &playHistory, nil
}
