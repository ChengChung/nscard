package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/chengchung/nscard/proto"
)

func GetPlayHistory(cred *CacheAccessToken) (*proto.UserPlayHistory, error) {
	request, err := http.NewRequest("GET", "https://news-api.entry.nintendo.co.jp/api/v1.2/users/me/play_histories", nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", cred.TokenType+" "+cred.AccessToken)
	request.Header.Set("User-Agent", user_agent)
	fmt.Println(cred.TokenType, cred.AccessToken)

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

	return proto.ParsePlayHistory(bytes)
}

func GetUserDetail(cred *CacheAccessToken) error {
	request, err := http.NewRequest("GET", "https://api.accounts.nintendo.com/2.0.0/users/me", nil)
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", cred.TokenType+" "+cred.AccessToken)
	request.Header.Set("User-Agent", user_agent)

	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))
	return nil
}
