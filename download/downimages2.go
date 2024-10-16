package download

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

func downloadImage2(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download image: %s", url)
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.ReadFrom(resp.Body)
	return err
}

func parseHTML2(htmlContent string) []string {
	var imgURLs []string
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return imgURLs
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			isMenuItem := false
			var href string
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == "menu-item" {
					isMenuItem = true
				}
				if a.Key == "href" {
					href = a.Val
				}
			}
			if isMenuItem && (strings.HasSuffix(href, ".png") || strings.HasSuffix(href, ".jpg")) {
				imgURLs = append(imgURLs, href)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return imgURLs
}

func DownImages2() {

	jsonFile, err := os.Open("links.json")
	if err != nil {
		fmt.Printf("Failed to open links.json: %s\n", err)
		return
	}
	defer jsonFile.Close()

	var links []struct {
		Href string `json:"href"`
		Text string `json:"text"`
	}

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Printf("Failed to read links.json: %s\n", err)
		return
	}

	err = json.Unmarshal(jsonData, &links)
	if err != nil {
		fmt.Printf("Failed to parse links.json: %s\n", err)
		return
	}
	semaphore := make(chan struct{}, 10) 
	var wg sync.WaitGroup

	for _, link := range links {
		resp, err := http.Get("https://learn.lianglianglee.com/" + link.Href + "/assets/")
		if err != nil {
			fmt.Printf("Failed to fetch HTML content: %s\n", err)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Failed to read response body: %s\n", err)
			return
		}

		htmlContent := string(body)


		imgURLs := parseHTML2(htmlContent)
		fmt.Printf("Found %d image URLs\n", len(imgURLs))
		baseURL := "https://learn.lianglianglee.com"

		for _, imgURL := range imgURLs {
			fullURL := baseURL + imgURL
			fileName := filepath.Base(imgURL)
			filePath := filepath.Join(link.Text+"\\assets", fileName)
			
			dir := filepath.Dir(filePath)
			
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					fmt.Printf("Failed to create directory: %s\n", err)
					continue
				}
			}

			wg.Add(1)
			go func(url, path string) {
				defer wg.Done()
				semaphore <- struct{}{} 
				defer func() { <-semaphore }() 

				err := downloadImage2(url, path)
				if err != nil {
					fmt.Printf("Failed to download image %s: %s\n", url, err)
					return
				}

				fmt.Printf("Image downloaded successfully: %s\n", path)
			}(fullURL, filePath)
		}
	}

	wg.Wait()
}
