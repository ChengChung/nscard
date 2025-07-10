package httpview

import (
	"sync"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/cors"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/middleware/requestid"
)

var (
	irisApp   *iris.Application
	init_once sync.Once
)

func GetIrisApp() *iris.Application {
	init_once.Do(func() {
		irisApp = iris.New()
		irisApp.Logger().SetLevel("debug")
		irisApp.Logger().Debugf(`Log level set to "debug"`)

		irisApp.UseRouter(requestid.New())
		irisApp.Logger().Debugf("Using <UUID4> to identify requests")

		irisApp.UseRouter(recover.New())

		irisApp.UseRouter(cors.New().
			ExtractOriginFunc(cors.DefaultOriginExtractor).
			ReferrerPolicy(cors.NoReferrerWhenDowngrade).
			AllowOriginFunc(cors.AllowAnyOrigin).
			Handler())
	})

	return irisApp
}
