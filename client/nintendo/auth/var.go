package auth

import (
	"net"
	"net/http"
	"time"
)

const (
	//	My Nintendo
	client_id          = "5c38e31cd085304b"
	login_redirect_uri = "npf5c38e31cd085304b://auth"
	grant_type         = "urn:ietf:params:oauth:grant-type:jwt-bearer-session-token"

	//	Nintendo Switch Online
	// client_id          = "71b963c1b7b6d119"
	// login_redirect_uri = "npf71b963c1b7b6d119://auth"
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
