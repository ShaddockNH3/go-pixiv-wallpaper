package catch

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"

	"github.com/PuerkitoBio/goquery"
)

// 定义API响应的结构体，用于解析JSON
type PixivAPIResponse struct {
	Error   bool       `json:"error"`
	Message string     `json:"message"`
	Body    IllustBody `json:"body"`
}

type IllustBody struct {
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Urls      ImageUrls `json:"urls"`
	PageCount int       `json:"pageCount"`
}

type ImageUrls struct {
	Original string `json:"original"`
}

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"

// createRequest 创建一个带有必要请求头的HTTP请求
func createRequest(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", "https://www.pixiv.net/")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	return req, nil
}

// fetchPageWithGoquery 获取HTML页面并用goquery解析
func fetchPageWithGoquery(url string) (*goquery.Document, error) {
	req, err := createRequest("GET", url)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// downloadImage 下载图片并根据指定规则保存
func downloadImage(url, pid, date string) error {
	req, err := createRequest("GET", url)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	targetDir := filepath.Join("image", "today_image")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目录 %s 失败: %v", targetDir, err)
	}

	fileName := date + filepath.Ext(url)
	filePath := filepath.Join(targetDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	log.Printf("成功下载壁纸: %s", filePath)
	return nil
}

// RunWallpaperFinder 主函数，执行壁纸寻找和下载逻辑
func RunWallpaperFinder(screenWidth int, screenHeight int, todayDate string) {
	if screenWidth == 0 || screenHeight == 0 {
		log.Println("传入的屏幕分辨率无效")
		return
	}

	screenRatio := float64(screenWidth) / float64(screenHeight)
	log.Printf("将根据配置的分辨率 %dx%d (比例: %.2f) 来寻找壁纸", screenWidth, screenHeight, screenRatio)

	rankingURL := "https://www.pixiv.net/ranking.php?mode=daily&content=illust"
	doc, err := fetchPageWithGoquery(rankingURL)
	if err != nil {
		log.Fatalf("获取排行榜页面失败了: %v", err)
	}

	var pids []string
	doc.Find("section.ranking-item").Each(func(i int, s *goquery.Selection) {
		if len(pids) >= 50 {
			return
		}
		if pid, exists := s.Attr("data-id"); exists {
			pids = append(pids, pid)
		}
	})

	if len(pids) == 0 {
		log.Println("在排行榜上没有找到任何作品ID")
		return
	}

	for _, pid := range pids {
		log.Printf("正在向 API 查询作品 PID: %s 的数据...", pid)

		apiURL := fmt.Sprintf("https://www.pixiv.net/ajax/illust/%s", pid)
		req, err := createRequest("GET", apiURL)
		if err != nil {
			log.Printf("为 PID: %s 创建 API 请求失败: %v", pid, err)
			continue
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("请求 API 失败, PID: %s, 错误: %v", pid, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Printf("API 响应状态码异常: %d, PID: %s", resp.StatusCode, pid)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("读取 API 响应体失败, PID: %s, 错误: %v", pid, err)
			continue
		}

		var apiResponse PixivAPIResponse
		if err := json.Unmarshal(body, &apiResponse); err != nil {
			log.Printf("解析 API JSON 失败, PID: %s, 错误: %v", pid, err)
			continue
		}

		if apiResponse.Error {
			log.Printf("API 返回错误信息, PID: %s, 消息: %s", pid, apiResponse.Message)
			continue
		}

		illustData := apiResponse.Body
		imgW := float64(illustData.Width)
		imgH := float64(illustData.Height)
		originalImageURL := illustData.Urls.Original

		if imgW == 0 || imgH == 0 {
			continue
		}

		imgRatio := imgW / imgH
		log.Printf("API 数据解析成功, 图片尺寸: %.0fx%.0f, 比例是 %.2f", imgW, imgH, imgRatio)

		if illustData.PageCount > 1 {
			log.Printf("作品 %s 是多页作品 (共 %d 页), 跳过", pid, illustData.PageCount)
			continue
		}

		if math.Abs(imgRatio-screenRatio) <= 0.1 {
			log.Println("匹配成功, 准备下载...")

			if originalImageURL == "" {
				log.Printf("找到匹配的图片但链接为空, PID: %s", pid)
				continue
			}

			err = downloadImage(originalImageURL, pid, todayDate)
			if err != nil {
				log.Printf("下载图片失败了, PID: %s, 错误: %v", pid, err)
			}

			log.Println("任务完成！")
			return
		}
	}

	log.Println("找遍了前50名也没有找到合适的壁纸。")
}
