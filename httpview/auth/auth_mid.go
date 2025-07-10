package auth

import (
	"github.com/chengchung/nscard/cache"
	"github.com/chengchung/nscard/client/nintendo/auth"
	"github.com/chengchung/nscard/utils"
	"github.com/kataras/iris/v12"
)

func AuthUser(ctx iris.Context) {
	uid := ctx.GetHeader("X-User-ID")
	token := ctx.GetHeader("X-User-Token")

	if len(uid) == 0 || len(token) == 0 {
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.StopExecution()
		return
	}

	cache, ok := cache.GetAuthClient(uid)
	if !ok || cache == nil {
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.StopExecution()
		return
	}

	if token != cache.GetSessionTokenHash() {
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.StopExecution()
		return
	}

	setAuthInfo(ctx, uid, cache)

	ctx.Next()
}

func setAuthInfo(ctx iris.Context, uid string, cli *auth.AccessClient) {
	ctx.Values().Set("nscard.auth_uid", uid)
	ctx.Values().Set("nscard.auth_client", cli)
	ctx.Values().Set("nscard.is_admin", utils.IsAdmin(uid))
}

func GetAuthUID(ctx iris.Context) string {
	uid, ok := ctx.Values().Get("nscard.auth_uid").(string)
	if !ok || len(uid) == 0 {
		return ""
	}
	return uid
}

func GetAuthClient(ctx iris.Context) *auth.AccessClient {
	cli, ok := ctx.Values().Get("nscard.auth_client").(*auth.AccessClient)
	if !ok || cli == nil {
		return nil
	}
	return cli
}

func IsAdmin(ctx iris.Context) bool {
	isAdmin, ok := ctx.Values().Get("nscard.is_admin").(bool)
	if !ok {
		return false
	}
	return isAdmin
}
