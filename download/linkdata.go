package download

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"golang.org/x/net/html"
)

// LinkData 用于保存a标签的href值和文本内容
type LinkData struct {
	Href  string `json:"href"`
	Text  string `json:"text"`
}

func FindLink() {
	url := "https://learn.lianglianglee.com/"

	// 发送HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to fetch URL: %s", resp.Status)
	}

	// 解析HTML
	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatalf("Error parsing HTML: %v", err)
	}

	var links []LinkData

	// 查找class为"book-post"的div标签
	var findLinks func(*html.Node)
	findLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "book-post") {
			// 遍历div的子节点，查找ul标签
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				if child.Type == html.ElementNode && child.Data == "div" {
					for ul := child.FirstChild; ul != nil; ul = ul.NextSibling {
						if ul.Type == html.ElementNode && ul.Data == "ul" {
							// 遍历ul的子节点，查找li标签
							for li := ul.FirstChild; li != nil; li = li.NextSibling {
								if li.Type == html.ElementNode && li.Data == "li" {
									for a := li.FirstChild; a != nil; a = a.NextSibling {
										if a.Type == html.ElementNode && a.Data == "a" {
											href := getAttr(a, "href")
											text := getText(a)
											if href != "" && text != "" {
												links = append(links, LinkData{Href: href, Text: text})
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		// 递归查找
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			findLinks(child)
		}
	}

	findLinks(doc)

	// 保存到JSON文件
	jsonData, err := json.MarshalIndent(links, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling to JSON: %v", err)
	}

	err = ioutil.WriteFile("links.json", jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing JSON file: %v", err)
	}

	fmt.Println("JSON data saved to links.json")
}

// hasClass 检查元素是否有指定的class
func hasClass(n *html.Node, class string) bool {
	if n.Type != html.ElementNode {
		return false
	}
	for _, a := range n.Attr {
		if a.Key == "class" && a.Val == class {
			return true
		}
	}
	return false
}

// getAttr 获取元素的属性值
func getAttr(n *html.Node, attrName string) string {
	for _, a := range n.Attr {
		if a.Key == attrName {
			return a.Val
		}
	}
	return ""
}

// getText 获取元素的文本内容
func getText(n *html.Node) string {
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			text += c.Data
		}
	}
	return text
}