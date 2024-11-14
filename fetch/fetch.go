package fetch

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/transform"
)

// Fetch 获取网页内容
func Fetch(url string, timeout time.Duration) (string, error) {
	// 创建一个新的 Colly 爬虫
	c := colly.NewCollector(
		// 设置请求超时
		colly.Async(true),
	)

	// 设置请求超时
	c.SetRequestTimeout(timeout)

	// 设置 User-Agent 模拟浏览器
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"

	// 设置请求头中的 Cookie（如果有需要）
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Cookie", "__jsluid_h=382abac99be2999e55b92c39875af9e2; __jsl_clearance=1731595344.967|0|djiPf%2FlegL2y5Oy60KMb669en5M%3D")
		// 设置 Referer
		r.Headers.Set("Referer", url)
	})

	var content string

	// 设置请求回调函数来获取页面的 HTML 内容
	c.OnResponse(func(r *colly.Response) {
		// 获取整个响应的 HTML 内容
		content = string(r.Body)
	})

	// 错误处理回调
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("请求错误: %v\n", err)
	})

	// 设置随机请求延迟，防止频繁请求被检测
	// c.Limit(&colly.LimitRule{
	// 	DomainGlob:  "*",
	// 	RandomDelay: 2 * time.Second, // 每次请求之间随机延迟 2 秒
	// })

	// 开始抓取页面
	err := c.Visit(url)
	if err != nil {
		return "", fmt.Errorf("colly 错误: %v", err)
	}

	// 等待爬虫完成抓取
	c.Wait()

	// 处理并返回抓取的HTML内容
	utf8Content, err := determineEncoding(content) // 这里调用 `determineEncoding` 来处理抓取到的 HTML 内容
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
