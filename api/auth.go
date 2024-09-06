package api

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"strings"
	"time"
)

func randBase64(len int) (string, error) {
	bytes := make([]byte, len)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func getAuthParams() (*AuthParams, error) {
	state, err := randBase64(36)
	if err != nil {
		return nil, err
	}
	codeVerifier, err := randBase64(32)
	if err != nil {
		return nil, err
	}
	hashedStr := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hashedStr[:])

	return &AuthParams{
		state:         state,
		codeVerifier:  codeVerifier,
		codeChallenge: codeChallenge,
	}, nil
}

func (cli *AuthClient) GetMyNintendoLoginURL() (string, error) {
	loginPortal := "https://accounts.nintendo.com/connect/1.0.0/authorize"

	url, err := urlpkg.Parse(loginPortal)
	if err != nil {
		return "", err
	}

	query := url.Query()
	query.Set("state", cli.authParams.state)
	query.Set("redirect_uri", login_redirect_uri)
	query.Set("client_id", client_id)
	query.Set("scope", "openid user user.mii user.email user.links[].id")
	query.Set("response_type", "session_token_code")
	query.Set("session_token_code_challenge", cli.authParams.codeChallenge)
	query.Set("session_token_code_challenge_method", "S256")
	query.Set("theme", "login_form")
	url.RawQuery = query.Encode()

	return url.String(), nil
}

func (cli *AuthClient) ParseCallbackURL(callbackUrl string) error {
	url, err := urlpkg.Parse(callbackUrl)
	if err != nil {
		return err
	}

	fragment := url.Fragment
	parts := strings.Split(fragment, "&")
	query := make(urlpkg.Values)

	for _, part := range parts {
		keyValue := strings.Split(part, "=")
		query.Set(keyValue[0], keyValue[1])
	}

	session_token_code := query.Get("session_token_code")
	if len(session_token_code) == 0 {
		return fmt.Errorf("invalid callback url %s", callbackUrl)
	}

	cli.sessionTokenCode = session_token_code

	return nil
}

func (cli *AuthClient) GetSessionCode() error {
	form := urlpkg.Values{}
	form.Set("client_id", client_id)
	form.Set("session_token_code", cli.sessionTokenCode)
	form.Set("session_token_code_verifier", cli.authParams.codeVerifier)

	request, err := http.NewRequest("POST", "https://accounts.nintendo.com/connect/1.0.0/api/session_token", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", user_agent)

	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 >= 4 {
		errmsg := &ErrorMessage{}
		err := json.NewDecoder(resp.Body).Decode(errmsg)
		if err != nil {
			return err
		}
		return fmt.Errorf("error: %s, description: %s", errmsg.Error, errmsg.ErrorDescription)
	}

	result := &SessionCodeResponse{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return err
	}

	cli.sessionToken = result.SessionToken

	return nil
}

type SessionCodeResponse struct {
	SessionToken string `json:"session_token"`
	Code         string `json:"code"`
}

func (cli *AuthClient) GetAccessToken() error {
	requestBody := AccessTokenRequest{
		ClientID:     client_id,
		GrantType:    grant_type,
		SessionToken: cli.sessionToken,
	}

	jsonReq, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", "https://accounts.nintendo.com/connect/1.0.0/api/token", bytes.NewReader(jsonReq))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("User-Agent", user_agent)

	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 >= 4 {
		errmsg := &ErrorMessage{}
		err := json.NewDecoder(resp.Body).Decode(errmsg)
		if err != nil {
			return err
		}
		return fmt.Errorf("error: %s, description: %s", errmsg.Error, errmsg.ErrorDescription)
	}

	result := &AccessTokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return err
	}

	cli.cachetoken = &CacheAccessToken{
		TokenType:       result.TokenType,
		AccessToken:     result.AccessToken,
		ExpireTimestamp: time.Now().Add(time.Duration(result.ExpiresIn) * time.Second).Add(-20 * time.Second),
		IDToken:         result.IDToken,
		Scope:           result.Scope,
	}

	return nil
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

type ErrorMessage struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (cli *AuthClient) GetToken() (*CacheAccessToken, error) {
	if cli.cachetoken == nil || time.Now().After(cli.cachetoken.ExpireTimestamp) {
		if err := cli.GetAccessToken(); err != nil {
			return nil, err
		}
	}
	return cli.cachetoken, nil
}
