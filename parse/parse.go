package parse

import (
	"fmt"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
)

type Result struct {
	Title    string
	Endpoint string
	Date     time.Time
}

// Parse 解析给定的 HTML 内容并提取标题和链接
func Parse(htmlContent string) (*Result, error) {
	// results := []Result{}

	result := &Result{}

	// 使用 soup 解析 HTML 内容
	doc := soup.HTMLParse(htmlContent)

	// 提取所有 tiles-item-text 元素
	paragraphs := doc.FindAll("div", "class", "tiles-item-text")
	for _, paragraph := range paragraphs {
		// result := Result{}

		// 找到日期并去除空白字符
		dateStr := strings.TrimSpace(paragraph.Find("div", "class", "tiles-item-text-date").Text())
		date, err := time.Parse("January 2, 2006", dateStr) // 解析日期
		if err != nil {
			return nil, fmt.Errorf("日期解析错误: %v", err)
		}
		result.Date = date.AddDate(0, 0, 1) // 加一天

		// 获取当前时间
		// now := time.Now()
		// 设置对比时间为 2024年10月24日
		compareDate, err := time.Parse("2006年01月02日", "2024年10月24日")
		if err != nil {
			return nil, fmt.Errorf("对比日期解析错误: %v", err)
		}

		if result.Date.Before(compareDate) {
			return nil, nil
		}

		// 找到标题和链接
		title := paragraph.Find("h3", "class", "tiles-item-text-title")
		if title.Error == nil {
			result.Endpoint = title.Find("a").Attrs()["href"]
			result.Title = title.Find("a").Text()
		} else {
			return nil, fmt.Errorf("未找到标题: %v", title.Error)
		}

		// results = append(results, result)
	}

	// 打印结果
	// for _, res := range results {
	fmt.Printf("Title: %s\nEndpoint: %s\nDate: %s\n", result.Title, result.Endpoint, result.Date.Format("January 2, 2006"))
	// }

	return result, nil
}

// func getTitle(htmlContent string){}
