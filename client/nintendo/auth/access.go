package auth

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chengchung/nscard/client/nintendo"
	"github.com/sirupsen/logrus"
)

type AccessClient struct {
	sessionToken string

	lck           *sync.Mutex
	cond          *sync.Cond
	running       bool
	err           error
	cacheResponse atomic.Pointer[CacheAccessToken]
}

func (cli *AccessClient) GetSessionTokenHash() string {
	md5sum := md5.Sum([]byte(cli.sessionToken))
	md5hex := hex.EncodeToString(md5sum[:])
	return md5hex
}

type CacheAccessToken struct {
	TokenType       string
	AccessToken     string
	ExpireTimestamp time.Time
	IDToken         string
	Scope           []string
}

func (c *CacheAccessToken) getAuthHeader() (string, error) {
	if time.Now().After(c.ExpireTimestamp) {
		return "", fmt.Errorf("access token expired")
	}
	return c.TokenType + " " + c.AccessToken, nil
}

func NewAccessClient(sessionToken string) *AccessClient {
	lck := &sync.Mutex{}

	return &AccessClient{
		sessionToken: sessionToken,
		lck:          lck,
		cond:         sync.NewCond(lck),
	}
}

func (cli *AccessClient) GetAccessToken() (string, error) {
	cacheRsp := cli.cacheResponse.Load()

	if cacheRsp != nil {
		if header, err := cacheRsp.getAuthHeader(); err == nil {
			return header, nil
		}
	}

	cli.lck.Lock()

	recoverFromWaiting := false
	for cli.running {
		cli.cond.Wait()
		recoverFromWaiting = true
	}
	if recoverFromWaiting {
		err := cli.err
		cache := cli.cacheResponse.Load()
		cli.lck.Unlock()
		if err != nil {
			return "", err
		}
		if cache == nil {
			return "", fmt.Errorf("unexpected no access token available, please try again later")
		}
		return cache.getAuthHeader()
	}

	//	current goroutine is the one that will fetch the access token
	cli.running = true
	cli.lck.Unlock()

	cache, err := cli.fetchAccessToken()
	if err != nil {
		cli.cacheResponse.Store(nil)
	} else {
		cli.cacheResponse.Store(cache)
	}

	cli.lck.Lock()
	cli.err = err
	cli.running = false
	cli.cond.Broadcast()
	cli.lck.Unlock()

	if err != nil {
		return "", err
	} else {
		return cache.getAuthHeader()
	}
}

func (cli *AccessClient) fetchAccessToken() (*CacheAccessToken, error) {
	requestBody := AccessTokenRequest{
		ClientID:     client_id,
		GrantType:    grant_type,
		SessionToken: cli.sessionToken,
	}

	jsonReq, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", "https://accounts.nintendo.com/connect/1.0.0/api/token", bytes.NewReader(jsonReq))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("User-Agent", nintendo.UserAgent)

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 >= 4 {
		errmsg := &ErrorMessage{}
		err := json.NewDecoder(resp.Body).Decode(errmsg)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error: %s, description: %s", errmsg.Error, errmsg.ErrorDescription)
	}

	result := &AccessTokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return nil, err
	}

	cache := &CacheAccessToken{
		TokenType:       result.TokenType,
		AccessToken:     result.AccessToken,
		ExpireTimestamp: time.Now().Add(time.Duration(result.ExpiresIn) * time.Second).Add(-20 * time.Second),
		IDToken:         result.IDToken,
		Scope:           result.Scope,
	}

	logrus.Infof("fetched access token for session %s, expires in %d seconds", cli.GetSessionTokenHash(), result.ExpiresIn)

	return cache, nil
}

type AccessTokenRequest struct {
	ClientID     string `json:"client_id"`
	GrantType    string `json:"grant_type"`
	SessionToken string `json:"session_token"`
}

type AccessTokenResponse struct {
	TokenType   string   `json:"token_type"`
	ExpiresIn   int      `json:"expires_in"`
	AccessToken string   `json:"access_token"`
	IDToken     string   `json:"id_token"`
	Scope       []string `json:"scope"`
}

func (cli *AccessClient) WithAccessToken(fn func(auth_header string) error) error {
	auth, err := cli.GetAccessToken()
	if err != nil {
		return err
	}

	return fn(auth)
}

func WithAccessToken[T any](fn func(auth_header string) (T, error)) func(cli *AccessClient) (T, error) {
	return func(cli *AccessClient) (T, error) {
		var result T
		err := cli.WithAccessToken(func(auth_header string) error {
			var err error
			result, err = fn(auth_header)
			return err
		})
		if err != nil {
			return result, err
		}
		return result, nil
	}
}
