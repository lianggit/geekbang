//下载URL提供的所有页面，转为html保存，单个页面下载
package download

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	// "sync"
	"time"
	// "time"

	"golang.org/x/net/html"
)

// LinkData 用于保存a标签的href值和文本内容
type LinkData2 struct {
	Href  string `json:"href"`
	Text  string `json:"text"`
}
// Article 结构体用于存储解析后的href和id值
type Article struct {
	Href string `json:"href"`
	ID   string `json:"id"`
}

func CreateFolder() {
	// 读取JSON文件
	jsonData, err := ioutil.ReadFile("links.json")
	if err != nil {
		log.Fatalf("Error reading JSON file: %v", err)
	}

	var links []LinkData2

	// 解析JSON数据
	err = json.Unmarshal(jsonData, &links)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}

	// 为href添加前缀并创建文件夹
	for _, link := range links {
		
		folderName := link.Text

		// 检查文件夹是否已存在
		_, err := os.Stat(folderName)
		if os.IsNotExist(err) {
			// 文件夹不存在，创建它
			err = os.Mkdir(folderName, os.ModePerm)
			if err != nil {
				log.Printf("Error creating folder %s: %v", folderName, err)
			} else {
				fmt.Printf("Folder %s created successfully\n", folderName)
			}
		} else if err != nil {
			// 其他错误
			log.Printf("Error checking folder %s: %v", folderName, err)
		} else {
			// 文件夹已存在
			fmt.Printf("Folder %s already exists\n", folderName)
		// 如果文件夹已存在，跳过当前循环，继续下一个
		continue
		}
		link.Href = "https://learn.lianglianglee.com" + link.Href
		GetHref(link.Href, folderName)
	}
}


// findMenuItemAnchors 查找所有具有特定类名的<a>标签并提取href和id属性
func findMenuItemAnchors(n *html.Node, targetClass string) ([]Article, error) {
	var articles []Article

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			// 查找class属性
			var class string
			for _, a := range n.Attr {
				if a.Key == "class" {
					class = a.Val
					break
				}
			}
			// 如果类名匹配，提取href和id属性
			if class == targetClass {
				var href, id string
				for _, a := range n.Attr {
					if a.Key == "href" {
						href = "https://learn.lianglianglee.com" +  a.Val
					} else if a.Key == "id" {
						id = a.Val
					}
				}
				if href != "" {
					articles = append(articles, Article{Href: href, ID: id})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)

	return articles, nil
}

func GetHref(url string, folderName string) {
	// 发送HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 解析HTML
	doc, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	// 查找所有具有特定类名的<a>标签
	targetClass := "menu-item"
	articles, err := findMenuItemAnchors(doc, targetClass)
	if err != nil {
		panic(err)
	}

	// 将结果保存为JSON
	jsonData, err := json.MarshalIndent(articles, "", "  ")
	if err != nil {
		panic(err)
	}

	// 保存JSON数据到文件
	err = ioutil.WriteFile(folderName+"/menu_items.json", jsonData, 0644)
	if err != nil {
		panic(err)
	}
	Download(folderName)
	fmt.Println(string(jsonData))
	fmt.Println("JSON data has been written to menu_items.json")
}
// MenuItem 定义了JSON文件中的条目结构
type MenuItem struct {
	Href string `json:"href"`
	ID   string `json:"id"`
}

func Download(folderName string) {
	// Read JSON file
	jsonData, err := ioutil.ReadFile(folderName + "/menu_items.json")
	if err != nil {
		fmt.Printf("Error reading JSON file: %v\n", err)
		return
	}

	// Parse JSON data
	var items []MenuItem
	err = json.Unmarshal(jsonData, &items)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	delay := time.Second * 5 // 3 second delay between downloads

	for _, item := range items {
		fmt.Printf("Downloading: %s\n", item.Href)
		err := downloadFile(item.ID, item.Href, folderName)
		if err != nil {
			fmt.Printf("Download failed: %v\n", err)
		}

		time.Sleep(delay) // Wait for 3 seconds before the next download
	}
}

// downloadFile 函数用于下载文件并保存
func downloadFile(filename, url string, folderName string) error {
	maxRetries := 3
	retryDelay := time.Second * 5

	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("下载失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			if i < maxRetries-1 {
				fmt.Printf("遇到 429 错误，等待 %v 后重试...\n", retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("服务器持续返回错误状态码: 429")
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
		}

		// 创建文件，添加.html后缀
		file, err := os.Create(folderName + "/" + filename + ".html")
		if err != nil {
			return fmt.Errorf("创建文件失败: %v", err)
		}
		defer file.Close()

		// 将响应内容写入文件
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}

		return nil // 成功下载，退出循环
	}

	return fmt.Errorf("下载失败，超过最大重试次数")
}
