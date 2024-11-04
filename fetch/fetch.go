package fetch

import (
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
func Fetch(url string) (string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// 设置超时
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.OuterHTML("html", &htmlContent), // 获取整个 HTML 内容
	)
	if err != nil {
		return "", err
	}

	// 检测编码并转换
	utf8Body, err := determineEncoding(htmlContent)
	if err != nil {
		fmt.Printf(err.Error())
		return "", fmt.Errorf("wrong with: %w", err)
	}

	return utf8Body, nil
}

// determineEncoding 检测编码
func determineEncoding(htmlContent string) (string, error) {
	// 解析 HTML 文档
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	var encoding string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			for _, attr := range n.Attr {
				if attr.Key == "charset" {
					encoding = attr.Val // 提取 charset
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	// 如果未找到 charset，默认使用 UTF-8
	if encoding == "" {
		encoding = "utf-8"
	}

	// 获取对应的编码转换器
	decoder, err := htmlindex.Get(encoding)
	if err != nil {
		return "", err
	}

	// 转换为 UTF-8
	reader := transform.NewReader(strings.NewReader(htmlContent), decoder.NewDecoder())
	utf8Body, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(utf8Body), nil
}
