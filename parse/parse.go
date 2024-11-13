package parse

import (
	"fmt"
	"time"

	"github.com/anaskhan96/soup"

	"code/config"
)

type Result struct {
	Title    string
	Endpoint string
	Date     time.Time
}

// Parse 解析给定的 HTML 内容并提取标题和链接
func Parse(htmlContent string, siteConfig config.SiteConfig) (*Result, error) {
	result := &Result{}

	// 使用 soup 解析 HTML 内容
	doc := soup.HTMLParse(htmlContent)
	if doc.Error != nil {
		return nil, fmt.Errorf("HTML 解析错误: %v", doc.Error)
	}

	// 只抓取第一个符合 contentSelector 的元素
	contentSelector := siteConfig.ParseRules["content"]
	contentElement := doc.Find(contentSelector)
	if contentElement.Error != nil {
		return nil, fmt.Errorf("未找到符合内容选择器 (%s) 的元素", contentSelector)
	}

	// 解析日期
	dateStr := extractDate(contentElement, siteConfig.ParseRules["date"])
	if dateStr == "" {
		return nil, fmt.Errorf("未能提取日期")
	}

	// 打印日期
	fmt.Printf("时间：%s\n", dateStr)

	// 解析日期格式
	date, err := parseDate(dateStr, siteConfig.DateFormats)
	if err != nil {
		return nil, fmt.Errorf("日期解析错误: %v", err)
	}
	result.Date = date

	// 比较日期，如果日期小于设定的对比日期，则跳过
	compareDate, err := time.Parse("2006年01月02日", "2024年10月24日")
	if err != nil {
		return nil, fmt.Errorf("对比日期解析错误: %v", err)
	}
	if result.Date.Before(compareDate) {
		return nil, nil
	}

	// 获取标题
	title := extractTitle(contentElement, siteConfig.ParseRules["title_tag"], siteConfig.ParseRules["title_class"])
	if title == "" {
		return nil, fmt.Errorf("未找到标题")
	}
	result.Title = title
	fmt.Printf("标题：%s\n", title)

	// 获取链接
	endpoint := extractLink(contentElement, siteConfig.ParseRules["link_tag"], siteConfig.ParseRules["link_class"])
	if endpoint == "" {
		return nil, fmt.Errorf("未找到链接")
	}
	result.Endpoint = endpoint
	fmt.Printf("链接：%s\n", endpoint)

	// 打印结果
	fmt.Printf("Title: %s\nEndpoint: %s\nDate: %s\n", result.Title, result.Endpoint, result.Date.Format("January 2, 2006"))

	return result, nil
}

// extractDate 提取日期字符串，支持多个日期格式
func extractDate(element soup.Root, dateSelector string) string {
	// 假设您有日期的 CSS 选择器
	dateElement := element.Find(dateSelector)
	if dateElement.Error != nil {
		return ""
	}
	return dateElement.Text()
}

// parseDate 尝试使用多种日期格式来解析日期
func parseDate(dateStr string, dateFormats []string) (time.Time, error) {
	for _, format := range dateFormats {
		date, err := time.Parse(format, dateStr)
		if err == nil {
			return date, nil
		}
	}
	return time.Time{}, fmt.Errorf("日期解析失败，所有格式都无法解析")
}

// extractTitle 提取标题
func extractTitle(element soup.Root, titleTag, titleClass string) string {
	// 假设您有标题的标签和类名
	titleElement := element.Find(titleTag, "class", titleClass)
	if titleElement.Error != nil {
		return ""
	}
	return titleElement.Text()
}

// extractLink 提取文章链接
func extractLink(element soup.Root, linkTag, linkClass string) string {
	// 假设您有链接的标签和类名
	linkElement := element.Find(linkTag, "class", linkClass)
	if linkElement.Error != nil {
		return ""
	}
	return linkElement.Attrs()["href"]
}
