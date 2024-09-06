package api

import (
	"net"
	"net/http"
	"time"
)

var (
	client = &http.Client{
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
)

type AuthParams struct {
	state         string
	codeVerifier  string
	codeChallenge string
}

type AuthClient struct {
	authParams       *AuthParams
	sessionTokenCode string
	sessionToken     string

	cachetoken *CacheAccessToken
}

type CacheAccessToken struct {
	TokenType       string
	AccessToken     string
	ExpireTimestamp time.Time
	IDToken         string
	Scope           []string
}

func NewAuthClient() (*AuthClient, error) {
	authParams, err := getAuthParams()
	if err != nil {
		return nil, err
	}
	return &AuthClient{authParams: authParams}, nil
}
