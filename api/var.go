package api

const (
	//	My Nintendo
	client_id          = "5c38e31cd085304b"
	login_redirect_uri = "npf5c38e31cd085304b://auth"
	grant_type         = "urn:ietf:params:oauth:grant-type:jwt-bearer-session-token"

	//	Nintendo Switch Online
	// client_id          = "71b963c1b7b6d119"
	// login_redirect_uri = "npf71b963c1b7b6d119://auth"

	app_version = "2.2.4"
	user_agent  = "com.nintendo.znej/" + app_version + " (Android/14.0)"
)
