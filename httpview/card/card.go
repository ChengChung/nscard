package card

import (
	"errors"
	"time"

	"github.com/chengchung/nscard/db"
	"github.com/chengchung/nscard/httpview"
	"github.com/chengchung/nscard/httpview/common"
	"github.com/chengchung/nscard/task/card"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

func init() {
	party := httpview.GetIrisApp().Party("/card/")

	party.Get("/uid/{userid}", get_card_img)
}

var renderConcurrency = make(chan struct{}, 10) // Limit concurrent rendering to 10

func get_card_img(ctx iris.Context) {
	userid := ctx.Params().GetString("userid")
	if userid == "" {
		ctx.JSON(common.HandleError(errors.New("user ID is required")))
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	cache, err := db.GetSwitchCardCache(userid)
	if err != nil && err != gorm.ErrRecordNotFound {
		ctx.JSON(common.HandleError(errors.New("failed to retrieve card cache")))
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}
	if err == gorm.ErrRecordNotFound {
		select {
		case renderConcurrency <- struct{}{}:
			err = card.RenderCard(userid)
			<-renderConcurrency
			if err != nil {
				ctx.JSON(common.HandleError(errors.New("failed to render card")))
				ctx.StatusCode(iris.StatusInternalServerError)
				return
			}
			cache, err = db.GetSwitchCardCache(userid)
		case <-time.After(5 * time.Second):
			ctx.JSON(common.HandleError(errors.New("rendering card timed out")))
			ctx.StatusCode(iris.StatusRequestTimeout)
			return
		}
	}

	if cache == nil {
		ctx.JSON(common.HandleError(errors.New("card not found")))
		ctx.StatusCode(iris.StatusNotFound)
		return
	}

	ctx.ContentType("image/svg+xml")
	ctx.Write([]byte(cache.Cache))
}
