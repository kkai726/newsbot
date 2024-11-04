package main

import (
	"code/fetch"
	"code/lark"
	"code/parse"
	"fmt"
)

const LarkWebHook = "https://open.feishu.cn/open-apis/bot/v2/hook/4329e228-0cab-499c-9915-b7dad761d1d6"

func main() {
	url := "https://nvidianews.nvidia.com/"
	utf8Body, err := fetch.Fetch(url)
	if err != nil {
		panic(err)
	}

	// fmt.Printf(utf8Body)
	result, err := parse.Parse(utf8Body)
	if err != nil {
		fmt.Printf(err.Error())
	}

	// 使用 fmt.Sprintf 创建消息
	message := fmt.Sprintf("最新消息: %s\n链接: %s\n日期: %s", result.Title, result.Endpoint, result.Date)

	err = lark.PushToLark(LarkWebHook, message)
	if err != nil {
		fmt.Printf(err.Error())
	}

	fmt.Println(result)
}
