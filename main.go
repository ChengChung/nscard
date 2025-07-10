package main

import (
	"flag"

	"github.com/chengchung/nscard/db"
	"github.com/chengchung/nscard/httpview"
	"github.com/chengchung/nscard/task"
	"github.com/chengchung/nscard/utils"
	"github.com/kataras/iris/v12"

	_ "github.com/chengchung/nscard/httpview/viewimport"
	_ "github.com/chengchung/nscard/task/taskimport"
)

func main() {
	flag.Parse()
	ParseConfig()

	err := db.Init(cfg.DBConfig)
	if err != nil {
		panic(err)
	}

	utils.InitAdminConfig(cfg.AdminConfig)

	task.Init(cfg.TaskConfigs)

	err = httpview.GetIrisApp().Run(iris.Addr(cfg.HttpConfig.ListenAddr))
	if err != nil {
		panic(err)
	}

	task.Stop()
}
