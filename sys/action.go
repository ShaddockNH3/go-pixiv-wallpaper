// 文件路径: sys/actions.go
package sys

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
)

// GetCurrentWallpaper 读取当前桌面壁纸并保存到指定路径
func GetCurrentWallpaper() (string, error) {
	// 从注册表读取当前壁纸的路径
	currentWallpaperPath, _, err := regist.GetStringValue("Wallpaper")
	if err != nil {
		return "", err
	}

	// 定义保存的目标文件夹
	destDir := "image/raw_image"
	// 确保目标文件夹存在
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}

	// 打开源文件
	srcFile, err := os.Open(currentWallpaperPath)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	// 创建目标文件
	destPath := filepath.Join(destDir, filepath.Base(currentWallpaperPath))
	destFile, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	// 复制文件内容
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return "", err
	}

	return destPath, nil
}

// SetTodayWallpaper 将 image/today_image 或 image/raw_image 下的第一张图设置为壁纸（居中）
func SetTodayWallpaper(path string) error {
	srcDir := path

	// 读取目录下的所有文件
	files, err := os.ReadDir(srcDir)
	if err != nil {
		log.Printf("错误：无法读取文件夹 %s: %v\n", srcDir, err)
		return err
	}

	var firstImagePath string
	// 找到第一个文件（不是文件夹）
	for _, file := range files {
		if !file.IsDir() {
			firstImagePath = filepath.Join(srcDir, file.Name())
			break
		}
	}

	// 如果没有找到图片，就返回一个错误
	if firstImagePath == "" {
		return errors.New("no image found in " + srcDir)
	}

	// 调用之前的函数来设置壁纸，使用 Fill 样式
	return SetDesktopWallpaper(firstImagePath, Fill)
}
