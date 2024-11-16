package parse

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/anaskhan96/soup"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tmt "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tmt/v20180321"

	"code/config"
)

type Result struct {
	Title    string
	Endpoint string
	Date     time.Time
}

func translate(text, targetLang string) (string, error) {
	// 使用你的腾讯云 SecretId 和 SecretKey
	// 加载配置文件
	config, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	credential := common.NewCredential(
		config.TencentParams.SecretID,  // 替换为你的 SecretId
		config.TencentParams.SecretKey, // 替换为你的 SecretKey
	)

	// 配置客户端的 region 和端点
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "tmt.tencentcloudapi.com" // 腾讯云翻译 API 的端点

	// 实例化 TMT 客户端
	client, err := tmt.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		return "", fmt.Errorf("创建 TMT 客户端失败: %v", err)
	}

	// 实例化请求对象
	request := tmt.NewTextTranslateRequest()

	// 获取项目 ID，假设没有配置其他项目，使用默认项目 ID: 0
	projectID := int64(0) // 默认项目 ID

	// 设置源文本、源语言和目标语言
	request.SourceText = common.StringPtr(text)
	request.Source = common.StringPtr("auto")
	request.Target = common.StringPtr(targetLang)
	request.ProjectId = &projectID

	// 发送请求并获取响应
	response, err := client.TextTranslate(request)
	if err != nil {
		if sdkErr, ok := err.(*errors.TencentCloudSDKError); ok {
			log.Printf("API 错误：%s", sdkErr)
		}
		return text, fmt.Errorf("翻译请求失败: %v", err)
	}

	// 提取翻译结果并返回
	return *response.Response.TargetText, nil
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
		paragraphs = append(paragraphs, doc.FindAll(siteConfig.ParseRules["content_tag"], siteConfig.ParseRules["content_mode"], className)...)
	}

	// fmt.Printf("内容是： %v\n", paragraphs[0].HTML())

	// 如果找不到匹配的内容，返回错误
	if len(paragraphs) == 0 {
		return nil, fmt.Errorf("未找到符合内容选择器 (%s) 的元素", siteConfig.ParseRules["content"])
	}

	// 提取日期并去除空白字符
	dateTags := strings.Split(siteConfig.ParseRules["date_tag"], ",") // 分割多个类名
	var dateStr string
	var dateElement soup.Root
	if siteConfig.ParseRules["date_in"] == "yes" {
		if siteConfig.ParseRules["date_mode"] != "" {
			dateElement = paragraphs[0].Find(siteConfig.ParseRules["date_tag"], siteConfig.ParseRules["date_mode"], siteConfig.ParseRules["date"])
		} else if siteConfig.ParseRules["date_mode"] == "" {
			dateElement = paragraphs[0]
			for _, dateTag := range dateTags {
				dateElement = dateElement.Find(dateTag)
			}
		}
	} else {
		dateElement = doc.Find(siteConfig.ParseRules["date_tag"], siteConfig.ParseRules["date_mode"], siteConfig.ParseRules["date"])

	}

	if dateElement.Error != nil {
		return nil, fmt.Errorf("未找到日期元素: %v", dateElement.Error)
	}
	dateStr = dateElement.Text()

	if dateStr != "" {
		resDate := strings.TrimSpace(dateStr)
		// 解析日期
		date, err := time.Parse(siteConfig.DateFormats[0], resDate) // 使用配置的第一个日期格式
		if err != nil {
			return nil, fmt.Errorf("日期解析错误: %v", err)
		}
		// 设置解析后的日期
		result.Date = date

		// 获取当前时间并进行对比
		compareDate, err := time.Parse("2006年01月02日", "2024年10月24日")
		if err != nil {
			return nil, fmt.Errorf("对比日期解析错误: %v", err)
		}
		if result.Date.Before(compareDate) {
			return nil, nil // 如果日期早于对比日期则跳过
		}
	}
	// 提取标题和链接
	var titleElement soup.Root
	if siteConfig.ParseRules["title_mode"] == "class" {
		// 处理 title_class 配置，分割多个类名
		titleClasses := strings.Split(siteConfig.ParseRules["title"], ",") // 分割多个类名
		for _, className := range titleClasses {
			// 查找标题元素
			tempElement := paragraphs[0].Find(siteConfig.ParseRules["title_tag"], "class", className)
			if tempElement.Error == nil {
				titleElement = tempElement
				break // 找到标题元素，跳出循环
			}
		}
	} else if siteConfig.ParseRules["title_mode"] == "" {
		titleElement = paragraphs[0]
	}

	// 如果没有找到标题，返回错误
	if titleElement.Error != nil {
		return nil, fmt.Errorf("未找到标题 %v", titleElement.Error)
	}

	// 提取标题和链接
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

	// 将标题翻译成中文
	translatedTitle, err := translate(result.Title, "zh")
	if err != nil {
		fmt.Printf("标题翻译失败: %v", err)
	}
	result.Title = translatedTitle

	// 拼接成完整的 URL（如果是相对路径）
	parsedURL, err := url.Parse(relativeURL)
	if err != nil {
		return nil, fmt.Errorf("链接解析错误: %v", err)
	}

	// 如果链接是相对的，则与 base_url 拼接
	if !parsedURL.IsAbs() {
		var fullURL string
		if siteConfig.RealURL == "" {
			fullURL, _ = getFullURL(siteConfig.BaseURL, relativeURL)
		} else {
			fullURL = siteConfig.RealURL + relativeURL
		}
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

		// 检查 relativeURL 是否有路径杂质，如果有则清理
		relativePath := parsedRelativeURL.Path
		relativePath = cleanRelativePath(relativePath)

		// 获取 base_url 和 relativeURL 的路径部分
		basePath := strings.TrimRight(parsedBaseURL.Path, "/")
		relativePath = strings.TrimLeft(relativePath, "/")

		// 拼接 basePath 和 relativePath
		fullPath := basePath + "/" + relativePath

		// 去除重复的部分
		// 以 basePath 的结尾为基准，去除重复的路径片段
		// if strings.HasPrefix(fullPath, basePath+"/") {
		// 	// fullPath = strings.TrimPrefix(fullPath, basePath)
		// 	fullPath = strings.Replace(fullPath, basePath, "", 1) // 只替换第一次出现的 basePath
		// }

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
