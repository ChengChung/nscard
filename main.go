package main

import (
	"fmt"

	"github.com/chengchung/nscard/api"
	"github.com/chengchung/nscard/games"
)

func main() {
	titles, err := games.GetTitleList(games.OptHardSwitch)
	if err != nil {
		fmt.Printf("获取游戏列表失败: %v\n", err)
		return
	}
	fmt.Printf("获取游戏列表成功，共 %d 个游戏\n", len(titles))
	for idx, title := range titles {
		if idx%1000 == 0 {
			fmt.Println(title)
		}
	}

	client, err := api.NewAuthClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	url, err := client.GetMyNintendoLoginURL()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(url)

	fmt.Println("请输入回调地址：")
	var callback string
	fmt.Scanln(&callback)
	if err := client.ParseCallbackURL(callback); err != nil {
		fmt.Println(err)
		return
	}

	if err := client.GetSessionCode(); err != nil {
		fmt.Println(err)
		return
	}

	cred, err := client.GetToken()
	if err != nil {
		fmt.Println(err)
		return
	}

	history, err := api.GetPlayHistory(cred)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(history)

	if err := api.GetUserDetail(cred); err != nil {
		fmt.Println(err)
		return
	}
}
