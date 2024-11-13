package parse

import (
	"errors"
	"fmt"
	"strings"
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

	// 解析内容部分
	contentSelector := siteConfig.ParseRules["content"]
	paragraphs := doc.FindAll("div", "class", contentSelector)
	if len(paragraphs) == 0 {
		return nil, fmt.Errorf("未找到符合内容选择器 (%s) 的元素", contentSelector)
	}

	// 解析日期
	dateStr := extractDate(paragraphs, siteConfig.ParseRules["date"], siteConfig.DateFormats)
	if dateStr == "" {
		return nil, fmt.Errorf("未能提取日期")
	}

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

	// 获取标题和链接
	title := extractTitle(paragraphs, siteConfig.ParseRules["title"])
	if title == "" {
		return nil, fmt.Errorf("未找到标题")
	}
	result.Title = title

	// 获取链接
	endpoint := extractLink(paragraphs, siteConfig.ParseRules["title"])
	if endpoint == "" {
		return nil, fmt.Errorf("未找到链接")
	}
	result.Endpoint = endpoint

	// 打印结果
	fmt.Printf("Title: %s\nEndpoint: %s\nDate: %s\n", result.Title, result.Endpoint, result.Date.Format("January 2, 2006"))

	return result, nil
}

// extractDate 提取日期字符串，支持多个日期格式
func extractDate(paragraphs []soup.Root, dateSelector string, dateFormats []string) string {
	var dateStr string
	for _, paragraph := range paragraphs {
		dateStr = strings.TrimSpace(paragraph.Find("div", "class", dateSelector).Text())
		if dateStr != "" {
			return dateStr
		}
	}
	return ""
}

// parseDate 尝试使用多种日期格式来解析日期
func parseDate(dateStr string, dateFormats []string) (time.Time, error) {
	for _, format := range dateFormats {
		date, err := time.Parse(format, dateStr)
		if err == nil {
			return date, nil
		}
	}
	return time.Time{}, errors.New("日期解析失败，所有格式都无法解析")
}

// extractTitle 提取标题
func extractTitle(paragraphs []soup.Root, titleSelector string) string {
	for _, paragraph := range paragraphs {
		title := paragraph.Find("h3", "class", titleSelector)
		if title.Error == nil {
			return title.Text()
		}
	}
	return ""
}

// extractLink 提取文章链接
func extractLink(paragraphs []soup.Root, titleSelector string) string {
	for _, paragraph := range paragraphs {
		title := paragraph.Find("h3", "class", titleSelector)
		if title.Error == nil {
			return title.Find("a").Attrs()["href"]
		}
	}
	return ""
}
