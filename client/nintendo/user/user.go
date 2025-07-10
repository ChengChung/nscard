package user

import (
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/chengchung/nscard/client/nintendo"
)

var client = &http.Client{
	Transport: &http.Transport{
		ResponseHeaderTimeout: 5 * time.Second,

		MaxIdleConns:    100,
		IdleConnTimeout: 90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   2 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	},
	Timeout: 8 * time.Second,
}

type UserInfo struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Country  string `json:"country"`
	IconUri  string `json:"iconUri"`
}

func GetUserDetail(auth_header string) (*UserInfo, error) {
	request, err := http.NewRequest("GET", "https://api.accounts.nintendo.com/2.0.0/users/me", nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", auth_header)
	request.Header.Set("User-Agent", nintendo.UserAgent)

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var userInfo UserInfo
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}
