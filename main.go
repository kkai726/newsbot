package main

import (
	"code/config"
	"code/db" // 引入 Redis 相关的包
	"code/fetch"
	"code/lark"
	"code/parse"
	"fmt"
	"log"
	"time"
)

const LarkWebHook = "https://open.feishu.cn/open-apis/bot/v2/hook/6710fb77-c813-4d32-b4a4-7a890a4d76db"

func main() {
	// 连接 Redis
	client, err := db.NewDatabaseClient(db.RedisType)
	if err != nil {
		log.Fatalf("无法连接 Redis: %v", err)
	}

	// 加载配置文件
	config, err := config.LoadConfig("config/webconfig.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 调用 scheduleFetch 函数，设置每 2 分钟执行一次
	scheduleFetch(config, 2*time.Minute, client)
}

// scheduleFetch 每隔指定时间执行一次抓取和处理操作
func scheduleFetch(config *config.Config, interval time.Duration, client db.DatabaseClient) {
	// 设置一个定时器，每次触发间隔为 interval（例如 15 分钟）
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 首次执行任务
	ProcessSites(config, client)

	// 使用 for range 监听 ticker.C，避免手动使用 select{}
	for range ticker.C {
		log.Println("开始执行定时任务...")
		ProcessSites(config, client) // 执行网站抓取和处理操作
	}
}

// ProcessSites 遍历配置中的每个站点，抓取网页内容并发送到 Lark
func ProcessSites(config *config.Config, client db.DatabaseClient) {
	// 循环遍历配置文件中的每个站点
	for _, site := range config.Sites {
		// 获取网站的 BaseURL
		url := site.BaseURL

		// 使用 Fetch 函数获取网页内容
		utf8Body, err := fetch.Fetch(url, 30*time.Second)
		if err != nil {
			log.Printf("Error fetching URL %s: %v\n", url, err)
			continue // 如果抓取失败，继续下一个 URL
		}

		// fmt.Printf("%v", utf8Body)

		// 解析网页内容，使用 Colly 解析内容
		result, err := parse.Parse(utf8Body, site)
		if err != nil {
			log.Printf("Error parsing content from URL %s: %v\n", url, err)
			continue // 如果解析失败，继续下一个 URL
		}

		// 检查 Redis 中是否已有该站点的内容
		existingEndpoint, err := client.GetKey(site.Name) // 获取 Redis 中存储的值
		if err == nil && existingEndpoint == result.Endpoint {
			// 如果 Redis 中已有相同的 Endpoint，跳过处理
			log.Printf("跳过站点 %s, 因为内容已存在\n", site.Name)
			continue
		}

		// 使用 fmt.Sprintf 创建消息
		// message := fmt.Sprintf("%s\n最新消息:\n%s\n链接: %s\n日期: %s", site.Name, result.Title, result.Endpoint, result.Date.Format("2006-01-02"))
		// 使用 fmt.Sprintf 创建消息，添加分隔符和突出显示的格式
		message := fmt.Sprintf(
			"【%s】\n\n"+ // 网站名称，突出显示
				"📢 最新消息:\n"+ // 添加提醒符号
				"➡️ %s\n\n"+ // 标题，使用箭头突出显示
				"🔗 链接: %s\n"+ // 链接行
				"📅 日期: %s", // 日期行
			site.Name,
			result.Title,
			result.Endpoint,
			result.Date.Format("2006-01-02"),
		)

		// 发送消息到 Lark
		err = lark.PushToLark(LarkWebHook, message)
		if err != nil {
			log.Printf("Error sending message to Lark for URL %s: %v\n", url, err)
			continue // 如果发送失败，继续下一个 URL
		}

		// 将新的 Endpoint 存入 Redis
		err = client.SetKey(site.Name, result.Endpoint)
		if err != nil {
			log.Printf("Error saving data to Redis for site %s: %v\n", site.Name, err)
			continue
		}

		// 打印成功的结果
		log.Printf("Successfully processed URL: %s\n", url)
		fmt.Println(result)
	}
}
