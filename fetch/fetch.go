package fetch

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"log"

	"github.com/chromedp/chromedp"
	"golang.org/x/net/html"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/transform"
)

// Fetch 获取网页内容
func Fetch(url string, timeout time.Duration) (string, error) {
	// 设置远程 Chromium 实例的调试地址
	chromeURL := "http://localhost:9222" // 默认情况下 Docker 会暴露该端口

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 创建一个带有远程调试地址的上下文
	allocCtx, cancel := chromedp.NewRemoteAllocator(ctx, chromeURL)
	defer cancel()

	// 创建一个新的上下文
	ctx2, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 在该上下文中运行任务
	var htmlContent string
	err := chromedp.Run(ctx2,
		chromedp.Navigate(url),
		chromedp.OuterHTML("html", &htmlContent), // 获取整个 HTML 内容
	)
	if err != nil {
		return "", fmt.Errorf("chromedp 错误: %v", err)
	}

	// 检测并转换编码
	utf8Body, err := determineEncoding(htmlContent)
	if err != nil {
		log.Printf("编码检测失败: %v", err)
		return "", fmt.Errorf("编码转换失败: %w", err)
	}

	return utf8Body, nil
}

// determineEncoding 检测并转换 HTML 内容编码
func determineEncoding(htmlContent string) (string, error) {
	// 解析 HTML 文档并查找 meta charset
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("HTML 解析错误: %v", err)
	}

	// 获取 meta 中的 charset 属性
	encoding := findCharsetInMetaTags(doc)
	if encoding == "" {
		// 如果未找到 charset，则使用 UTF-8 作为默认编码
		encoding = "utf-8"
	}

	// 如果 charset 不是 UTF-8，则进行编码转换
	if encoding != "utf-8" {
		return convertToUTF8(htmlContent, encoding)
	}

	// 如果已经是 UTF-8 编码，直接返回
	return htmlContent, nil
}

// findCharsetInMetaTags 查找 meta 标签中的 charset 属性
func findCharsetInMetaTags(doc *html.Node) string {
	var encoding string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			for _, attr := range n.Attr {
				if attr.Key == "charset" {
					encoding = attr.Val
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return encoding
}

// convertToUTF8 将 HTML 内容从指定编码转换为 UTF-8
func convertToUTF8(htmlContent, encoding string) (string, error) {
	// 获取指定编码的转换器
	decoder, err := htmlindex.Get(encoding)
	if err != nil {
		return "", fmt.Errorf("获取编码转换器失败: %v", err)
	}

	// 将内容转换为 UTF-8
	reader := transform.NewReader(strings.NewReader(htmlContent), decoder.NewDecoder())
	utf8Body, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("读取转换内容失败: %v", err)
	}

	return string(utf8Body), nil
}
