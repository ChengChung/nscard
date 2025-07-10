package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"strings"

	"github.com/chengchung/nscard/client/nintendo"
)

type LoginClient struct {
	state         string
	codeVerifier  string
	codeChallenge string
}

func randBase64(len int) (string, error) {
	bytes := make([]byte, len)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func NewLoginClient() (*LoginClient, error) {
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
	return &LoginClient{
		state:         state,
		codeVerifier:  codeVerifier,
		codeChallenge: codeChallenge,
	}, nil
}

func (cli *LoginClient) GetMyNintendoLoginURL() (string, string, error) {
	loginPortal := "https://accounts.nintendo.com/connect/1.0.0/authorize"

	url, err := urlpkg.Parse(loginPortal)
	if err != nil {
		return "", "", err
	}

	query := url.Query()
	query.Set("state", cli.state)
	query.Set("redirect_uri", login_redirect_uri)
	query.Set("client_id", client_id)
	query.Set("scope", "openid user user.mii user.email user.links[].id")
	query.Set("response_type", "session_token_code")
	query.Set("session_token_code_challenge", cli.codeChallenge)
	query.Set("session_token_code_challenge_method", "S256")
	query.Set("theme", "login_form")
	url.RawQuery = query.Encode()

	return url.String(), cli.codeVerifier, nil
}

type SessionClient struct {
	sessionTokenCodeVerifier string
	sessionTokenCode         string
}

func NewSessionClient(callbackUrl string, codeVerifier string) (*SessionClient, error) {
	url, err := urlpkg.Parse(callbackUrl)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("invalid callback url %s", callbackUrl)
	}

	return &SessionClient{
		sessionTokenCodeVerifier: codeVerifier,
		sessionTokenCode:         session_token_code,
	}, nil
}

func (cli *SessionClient) GetSessionToken() (string, error) {
	form := urlpkg.Values{}
	form.Set("client_id", client_id)
	form.Set("session_token_code", cli.sessionTokenCode)
	form.Set("session_token_code_verifier", cli.sessionTokenCodeVerifier)

	request, err := http.NewRequest("POST", "https://accounts.nintendo.com/connect/1.0.0/api/session_token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", nintendo.UserAgent)

	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 >= 4 {
		errmsg := &ErrorMessage{}
		err := json.NewDecoder(resp.Body).Decode(errmsg)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("error: %s, description: %s", errmsg.Error, errmsg.ErrorDescription)
	}

	result := &SessionCodeResponse{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return "", err
	}

	return result.SessionToken, nil
}

type SessionCodeResponse struct {
	SessionToken string `json:"session_token"`
	Code         string `json:"code"`
}

type ErrorMessage struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}
