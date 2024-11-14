package parse

import (
	"fmt"
	"net/url"
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

	// 处理配置中的多个类名
	contentClasses := strings.Split(siteConfig.ParseRules["content"], ",") // 分割多个类名
	var paragraphs []soup.Root
	for _, className := range contentClasses {
		// 遍历并查找包含该类名的 div
		paragraphs = append(paragraphs, doc.FindAll("div", "class", className)...)
	}

	// 如果找不到匹配的内容，返回错误
	if len(paragraphs) == 0 {
		return nil, fmt.Errorf("未找到符合内容选择器 (%s) 的元素", siteConfig.ParseRules["content"])
	}

	// 提取日期并去除空白字符
	dateStr := strings.TrimSpace(paragraphs[0].Find(siteConfig.ParseRules["date_tag"], "class", siteConfig.ParseRules["date"]).Text())
	if dateStr == "" {
		return nil, fmt.Errorf("未找到日期")
	}
	// 解析日期
	date, err := time.Parse(siteConfig.DateFormats[0], dateStr) // 使用配置的第一个日期格式
	if err != nil {
		return nil, fmt.Errorf("日期解析错误: %v", err)
	}
	// result.Date = date.AddDate(0, 0, 1) // 日期加一天
	result.Date = date

	// 获取当前时间并进行对比
	compareDate, err := time.Parse("2006年01月02日", "2024年10月24日")
	if err != nil {
		return nil, fmt.Errorf("对比日期解析错误: %v", err)
	}
	if result.Date.Before(compareDate) {
		return nil, nil // 如果日期早于对比日期则跳过
	}

	// 提取标题和链接
	// 处理 title_class 配置，分割多个类名
	titleClasses := strings.Split(siteConfig.ParseRules["title"], ",") // 分割多个类名
	var titleElement soup.Root
	for _, className := range titleClasses {
		// 查找标题元素
		tempElement := paragraphs[0].Find(siteConfig.ParseRules["title_tag"], "class", className)
		if tempElement.Error == nil {
			titleElement = tempElement
			break // 找到标题元素，跳出循环
		}
	}

	// 如果没有找到标题，返回错误
	if titleElement.Error != nil {
		return nil, fmt.Errorf("未找到标题: %v", titleElement.Error)
	}

	// 提取标题和链接
	// result.Title = titleElement.Text()
	// fmt.Printf("标题是：%v\n", result.Title)

	// 尝试从 <a> 标签中提取链接
	aElement := titleElement.Find("a")
	var relativeURL string

	if aElement.Error != nil {
		// 如果没有找到 <a> 标签，检查 titleElement 是否有 href 属性
		hrefAttr, ok := titleElement.Attrs()["href"]
		if !ok || hrefAttr == "" {
			return nil, fmt.Errorf("未找到链接")
		}
		relativeURL = hrefAttr
		result.Title = titleElement.Text()
	} else {
		// 如果找到了 <a> 标签，提取 href 属性
		hrefAttr, ok := aElement.Attrs()["href"]
		if !ok || hrefAttr == "" {
			return nil, fmt.Errorf("未找到链接")
		}
		relativeURL = hrefAttr
		result.Title = aElement.Text()
	}

	// fmt.Printf("链接是： %v\n", relativeURL)

	// 拼接成完整的 URL（如果是相对路径）
	parsedURL, err := url.Parse(relativeURL)
	if err != nil {
		return nil, fmt.Errorf("链接解析错误: %v", err)
	}

	// 如果链接是相对的，则与 base_url 拼接
	if !parsedURL.IsAbs() {
		fullURL, _ := getFullURL(siteConfig.BaseURL, relativeURL)
		result.Endpoint = fullURL

	} else {
		result.Endpoint = relativeURL // 如果是完整 URL，则直接使用
	}

	// 打印结果
	fmt.Printf("Title: %s\nEndpoint: %s\nDate: %s\n", result.Title, result.Endpoint, result.Date.Format("January 2, 2006"))

	return result, nil
}

// 去除 URL 中重复的路径部分
func getFullURL(baseURL, relativeURL string) (string, error) {
	// 解析 base_url 和 relativeURL
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("base_url 解析错误: %v", err)
	}

	parsedRelativeURL, err := url.Parse(relativeURL)
	if err != nil {
		return "", fmt.Errorf("relativeURL 解析错误: %v", err)
	}

	// 如果 relativeURL 是相对路径，拼接 base_url 和 relativeURL
	if !parsedRelativeURL.IsAbs() {
		// 获取 base_url 和 relativeURL 的路径部分
		basePath := strings.TrimRight(parsedBaseURL.Path, "/")
		relativePath := strings.TrimLeft(parsedRelativeURL.Path, "/")

		// 拼接 basePath 和 relativePath
		fullPath := basePath + "/" + relativePath

		// 去除重复的部分
		// 以 basePath 的结尾为基准，去除重复的路径片段
		if strings.HasPrefix(fullPath, basePath+"/") {
			fullPath = strings.TrimPrefix(fullPath, basePath)
		}

		// 拼接完整的 URL
		return parsedBaseURL.Scheme + "://" + parsedBaseURL.Host + fullPath, nil
	}

	// 如果 relativeURL 是完整的 URL，直接返回
	return relativeURL, nil
}
