package user

import (
	"errors"

	"github.com/chengchung/nscard/cache"
	"github.com/chengchung/nscard/client/nintendo/user"
	"github.com/chengchung/nscard/httpview"
	"github.com/chengchung/nscard/httpview/auth"
	"github.com/chengchung/nscard/httpview/common"
	"github.com/kataras/iris/v12"

	authpkg "github.com/chengchung/nscard/client/nintendo/auth"
)

func init() {
	party := httpview.GetIrisApp().Party("/user/")

	party.Use(auth.AuthUser)

	party.Get("/info", get_user_info)
	party.Post("/settings", post_user_settings)
	party.Get("/game_list", get_user_game_list)
}

func get_user_info(ctx iris.Context) {
	userId := auth.GetAuthUID(ctx)
	info, _ := cache.GetUserInfo(userId)

	ctx.JSON(info)
	ctx.StatusCode(iris.StatusOK)
}

func post_user_settings(ctx iris.Context) {
	userId := auth.GetAuthUID(ctx)
	//	update DB
	//	update cache
	var userSettings cache.UserCustomSettings
	if err := ctx.ReadJSON(&userSettings); err != nil {
		ctx.JSON(common.HandleError(errors.New("invalid request body")))
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	if err := cache.UpdateUserSettings(userId, &userSettings); err != nil {
		ctx.JSON(common.HandleError(err))
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.JSON(common.HandleSuccess("settings updated successfully"))
	ctx.StatusCode(iris.StatusOK)
}

func get_user_game_list(ctx iris.Context) {
	client := auth.GetAuthClient(ctx)

	history, err := authpkg.WithAccessToken(user.GetPlayHistory)(client)
	if err != nil {
		ctx.JSON(common.HandleError(err))
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	games := make([]string, 0, len(history.PlayHistories))
	for _, game := range history.PlayHistories {
		games = append(games, game.TitleName)
	}

	ctx.JSON(map[string]any{
		"games": games,
	})
	ctx.StatusCode(iris.StatusOK)
}
