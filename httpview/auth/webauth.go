package auth

import (
	"net/http"

	"github.com/chengchung/nscard/cache"
	"github.com/chengchung/nscard/client/nintendo/auth"
	"github.com/chengchung/nscard/client/nintendo/user"
	"github.com/chengchung/nscard/httpview"
	"github.com/chengchung/nscard/httpview/common"
	"github.com/kataras/iris/v12"
)

func init() {
	party := httpview.GetIrisApp().Party("/auth")

	party.Get("/login", get_login_page)
	party.Post("/login", post_login_page)
}

type loginParams struct {
	RedirectURI  string `json:"redirect_uri"`
	CodeVerifier string `json:"code_verifier"`
}

func get_login_page(ctx iris.Context) {
	var err error
	defer func() {
		if err != nil {
			ctx.JSON(common.HandleError(err))
			ctx.StatusCode(http.StatusInternalServerError)
		}
	}()

	cli, err := auth.NewLoginClient()
	if err != nil {
		return
	}

	url, cv, err := cli.GetMyNintendoLoginURL()
	if err != nil {
		return
	}
	ctx.JSON(loginParams{
		RedirectURI:  url,
		CodeVerifier: cv,
	})
	ctx.StatusCode(http.StatusOK)
}

type loginPostParams struct {
	CallbackUrl  string `json:"callback_url"`
	CodeVerifier string `json:"code_verifier"`
}

func post_login_page(ctx iris.Context) {
	var err error
	defer func() {
		if err != nil {
			ctx.JSON(common.HandleError(err))
			ctx.StatusCode(http.StatusInternalServerError)
		}
	}()

	var params loginPostParams
	if err = ctx.ReadJSON(&params); err != nil {
		return
	}

	cli, err := auth.NewSessionClient(params.CallbackUrl, params.CodeVerifier)
	if err != nil {
		return
	}
	sessionToken, err := cli.GetSessionToken()
	if err != nil {
		return
	}

	authcli := auth.NewAccessClient(sessionToken)

	var userInfo user.UserInfo
	err = authcli.WithAccessToken(func(auth_token string) error {
		info, err := user.GetUserDetail(auth_token)
		if err == nil {
			userInfo = *info
		}
		return err
	})
	if err != nil {
		return
	}

	if err = cache.CreateOrUpdateUserCache(&userInfo, sessionToken); err != nil {
		return
	}

	ctx.JSON(map[string]string{
		"token": authcli.GetSessionTokenHash(),
		"uid":   userInfo.ID,
	})
	ctx.StatusCode(http.StatusOK)
}
