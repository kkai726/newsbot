package fetch

import (
	"code/config"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"golang.org/x/net/html"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/transform"
)

// Fetch 获取网页内容
func Fetch(url string, timeout time.Duration, siteConfig config.SiteConfig) (string, error) {
	// 设置 WebSocket URL，指向远程 Chrome 实例
	wsURL := "ws://localhost:9222" // 连接到 WebSocket 上的 Chrome 实例

	// 创建远程上下文连接到远程 Chrome 实例
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 创建浏览器执行器分配器，并配置选项
	allocatorCtx, cancel := chromedp.NewRemoteAllocator(ctx, wsURL)
	defer cancel()

	// 创建新的浏览器上下文
	ctx, cancel = chromedp.NewContext(allocatorCtx)
	defer cancel()

	// 设置浏览器视口大小，模拟正常浏览器行为
	if err := chromedp.Run(ctx, chromedp.EmulateViewport(1280, 1024)); err != nil {
		return "", fmt.Errorf("设置浏览器视口错误: %v", err)
	}

	var content string
	// 执行浏览器操作，抓取网页内容
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url), // 导航到目标 URL
		chromedp.WaitVisible(siteConfig.ParseRules["content"], chromedp.ByQuery),                                     // 等待页面上的特定元素可见
		chromedp.Text(fmt.Sprintf("%s:first-of-type", siteConfig.ParseRules["content"]), &content, chromedp.ByQuery), // 获取特定元素的文本内容
	); err != nil {
		return "", fmt.Errorf("chromedp 错误: %v", err)
	}

	// 处理并返回抓取的 HTML 内容
	utf8Content, err := determineEncoding(content)
	if err != nil {
		return "", fmt.Errorf("编码转换错误: %v", err)
	}

	return utf8Content, nil
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
				if attr.Key == "http-equiv" && attr.Val == "Content-Type" {
					for _, metaAttr := range n.Attr {
						if metaAttr.Key == "content" && strings.Contains(metaAttr.Val, "charset") {
							encoding = strings.Split(metaAttr.Val, "charset=")[1]
							return
						}
					}
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
