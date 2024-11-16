package parse

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/anaskhan96/soup"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tmt "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tmt/v20180321"

	"code/config"
)

type Result struct {
	Title    string
	Endpoint string
	Date     time.Time
}

// translate 调用腾讯云翻译API，将文本翻译成目标语言
func translate(text, targetLang string) (string, error) {
	// 加载配置文件
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		return text, fmt.Errorf("加载配置失败: %v", err)
	}

	// 创建腾讯云客户端
	client, err := newTencentClient(cfg)
	if err != nil {
		return text, err
	}

	// 调用翻译API
	request := tmt.NewTextTranslateRequest()
	request.SourceText = common.StringPtr(text)
	request.Source = common.StringPtr("auto")
	request.Target = common.StringPtr(targetLang)
	request.ProjectId = common.Int64Ptr(0) // 默认项目ID

	response, err := client.TextTranslate(request)
	if err != nil {
		return text, fmt.Errorf("翻译请求失败: %v", err)
	}

	// 返回翻译后的文本
	return *response.Response.TargetText, nil
}

// newTencentClient 创建并配置腾讯云客户端
func newTencentClient(cfg *config.Config) (*tmt.Client, error) {
	credential := common.NewCredential(
		cfg.TencentParams.SecretID,
		cfg.TencentParams.SecretKey,
	)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "tmt.tencentcloudapi.com" // 腾讯云翻译API端点

	client, err := tmt.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		return nil, fmt.Errorf("创建 TMT 客户端失败: %v", err)
	}
	return client, nil
}

// Parse 解析HTML内容，提取标题、日期和链接
func Parse(htmlContent string, siteConfig config.SiteConfig) (*Result, error) {
	// 初始化结果结构体
	result := &Result{}

	// 解析HTML内容
	doc := soup.HTMLParse(htmlContent)
	if doc.Error != nil {
		return nil, fmt.Errorf("HTML解析错误: %v", doc.Error)
	}

	// 获取文章内容
	paragraphs, err := getContent(doc, siteConfig)
	if err != nil {
		return nil, err
	}

	// 提取日期
	date, err := getDate(paragraphs, doc, siteConfig)
	if err != nil {
		return nil, err
	}
	result.Date = date

	// 提取标题和链接
	title, endpoint, err := getTitleAndEndpoint(paragraphs, siteConfig)
	if err != nil {
		return nil, err
	}
	result.Title = title
	result.Endpoint = endpoint

	// 翻译标题
	translatedTitle, err := translate(result.Title, "zh")
	if err != nil {
		log.Printf("标题翻译失败: %v", err)
	}
	result.Title = translatedTitle

	// 返回结果
	fmt.Printf("Title: %s\nEndpoint: %s\nDate: %s\n", result.Title, result.Endpoint, result.Date.Format("January 2, 2006"))
	return result, nil
}

// getContent 获取文章内容
func getContent(doc soup.Root, siteConfig config.SiteConfig) ([]soup.Root, error) {
	contentClasses := strings.Split(siteConfig.ParseRules["content"], ",")
	var paragraphs []soup.Root
	for _, className := range contentClasses {
		paragraphs = append(paragraphs, doc.FindAll(siteConfig.ParseRules["content_tag"], siteConfig.ParseRules["content_mode"], className)...)
	}

	if len(paragraphs) == 0 {
		return nil, fmt.Errorf("未找到符合内容选择器 (%s) 的元素", siteConfig.ParseRules["content"])
	}
	return paragraphs, nil
}

// getDate 提取并解析文章日期
func getDate(paragraphs []soup.Root, doc soup.Root, siteConfig config.SiteConfig) (time.Time, error) {
	dateTags := strings.Split(siteConfig.ParseRules["date_tag"], ",")
	var dateElement soup.Root

	if siteConfig.ParseRules["date_in"] == "yes" {
		dateElement = findDateInParagraphs(paragraphs, dateTags)
	} else {
		dateElement = doc.Find(siteConfig.ParseRules["date_tag"], siteConfig.ParseRules["date_mode"], siteConfig.ParseRules["date"])
	}

	if dateElement.Error != nil {
		return time.Time{}, fmt.Errorf("未找到日期元素: %v", dateElement.Error)
	}

	dateStr := strings.TrimSpace(dateElement.Text())
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("日期为空")
	}

	// 解析日期
	date, err := time.Parse(siteConfig.DateFormats[0], dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("日期解析错误: %v", err)
	}

	// 日期对比
	compareDate, err := time.Parse("2006年01月02日", "2024年10月24日")
	if err != nil {
		return time.Time{}, fmt.Errorf("对比日期解析错误: %v", err)
	}

	if date.Before(compareDate) {
		return time.Time{}, nil // 如果日期早于对比日期，则跳过
	}

	return date, nil
}

// findDateInParagraphs 在文章中查找日期
func findDateInParagraphs(paragraphs []soup.Root, dateTags []string) soup.Root {
	var dateElement soup.Root
	for _, dateTag := range dateTags {
		dateElement = paragraphs[0].Find(dateTag)
		if dateElement.Error == nil {
			break
		}
	}
	return dateElement
}

// getTitleAndEndpoint 提取标题和链接
func getTitleAndEndpoint(paragraphs []soup.Root, siteConfig config.SiteConfig) (string, string, error) {
	var titleElement soup.Root
	var title string
	if siteConfig.ParseRules["title_mode"] == "class" {
		// 处理 title_class 配置
		titleClasses := strings.Split(siteConfig.ParseRules["title"], ",")
		for _, className := range titleClasses {
			titleElement = paragraphs[0].Find(siteConfig.ParseRules["title_tag"], "class", className)
			if titleElement.Error == nil {
				break
			}
		}
	} else if siteConfig.ParseRules["title_mode"] == "" {
		titleElement = paragraphs[0]
	}

	if titleElement.Error != nil {
		return "", "", fmt.Errorf("未找到标题 %v", titleElement.Error)
	}

	// 获取链接
	aElement := titleElement.Find("a")
	var relativeURL string
	if aElement.Error != nil {
		// 检查 titleElement 是否有 href 属性
		hrefAttr, ok := titleElement.Attrs()["href"]
		if !ok || hrefAttr == "" {
			return "", "", fmt.Errorf("未找到链接")
		}
		relativeURL = hrefAttr
		title = titleElement.Text()
	} else {
		hrefAttr, ok := aElement.Attrs()["href"]
		if !ok || hrefAttr == "" {
			return "", "", fmt.Errorf("未找到链接")
		}
		relativeURL = hrefAttr
		title = aElement.Text()
	}

	// 拼接完整URL
	parsedURL, err := url.Parse(relativeURL)
	if err != nil {
		return "", "", fmt.Errorf("链接解析错误: %v", err)
	}

	var fullURL string
	if !parsedURL.IsAbs() {
		if siteConfig.RealURL == "" {
			fullURL, _ = getFullURL(siteConfig.BaseURL, relativeURL)
		} else {
			fullURL = siteConfig.RealURL + relativeURL
		}
	} else {
		fullURL = relativeURL
	}

	return title, fullURL, nil
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

		// 检查 relativeURL 是否有路径杂质，如果有则清理
		relativePath := parsedRelativeURL.Path
		relativePath = cleanRelativePath(relativePath)

		// 获取 base_url 和 relativeURL 的路径部分
		basePath := strings.TrimRight(parsedBaseURL.Path, "/")
		relativePath = strings.TrimLeft(relativePath, "/")

		// 拼接 basePath 和 relativePath
		fullPath := basePath + "/" + relativePath

		// 如果 fullPath 中有多个 basePath 部分，去除多余的部分
		// 保证最终只保留一个 basePath
		fullPath = strings.Replace(fullPath, basePath, "", -1) // 去掉所有的 basePath 部分
		// 拼接最终的完整路径，保留一个 basePath
		fullPath = basePath + fullPath

		// 拼接完整的 URL
		return parsedBaseURL.Scheme + "://" + parsedBaseURL.Host + fullPath, nil
	}

	// 如果 relativeURL 是完整的 URL，直接返回
	return relativeURL, nil
}

// 清理相对路径中的杂质，只保留第一个 "/" 及其之后的部分
func cleanRelativePath(relativeURL string) string {
	// 查找第一个 "/" 的位置
	index := strings.Index(relativeURL, "/")
	if index == -1 {
		// 如果没有找到 "/", 就直接返回原路径
		return relativeURL
	}
	// 返回从第一个 "/" 开始的路径
	return relativeURL[index:]
}
