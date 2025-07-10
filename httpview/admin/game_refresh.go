package admin

import (
	"net/http"

	"github.com/chengchung/nscard/httpview"
	"github.com/chengchung/nscard/httpview/auth"
	"github.com/chengchung/nscard/httpview/common"
	"github.com/chengchung/nscard/task/games"
	"github.com/kataras/iris/v12"
)

func init() {
	party := httpview.GetIrisApp().Party("/admin")

	party.Use(auth.AuthUser, check_admin)

	party.Get("/refresh_jp_xml", refresh_jp_xml)
	party.Get("/refresh_jp_search", refresh_jp_search)
}

func check_admin(ctx iris.Context) {
	isAdmin := auth.IsAdmin(ctx)
	if !isAdmin {
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.StopExecution()
		return
	}

	ctx.Next()
}

func refresh_jp_xml(ctx iris.Context) {
	var err error
	defer func() {
		if err != nil {
			ctx.JSON(common.HandleError(err))
			ctx.StatusCode(iris.StatusInternalServerError)
			return
		}
	}()

	t := &games.RefreshJPTitlesXMLToDBTask{}
	err = t.Run()
	if err != nil {
		return
	}

	ctx.JSON(common.HandleSuccess("refresh jp store titles via xml success"))
	ctx.StatusCode(http.StatusOK)
}

func refresh_jp_search(ctx iris.Context) {
	var err error
	defer func() {
		if err != nil {
			ctx.JSON(common.HandleError(err))
			ctx.StatusCode(iris.StatusInternalServerError)
			return
		}
	}()

	t := &games.RefreshJPTitlesSearchToDBTask{}
	err = t.Run()
	if err != nil {
		return
	}

	ctx.JSON(common.HandleSuccess("refresh jp store titles via search api success"))
	ctx.StatusCode(http.StatusOK)
}
