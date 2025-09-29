// 文件路径: sys/actions.go
package sys

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// GetCurrentWallpaper 读取当前桌面壁纸并保存到指定路径
func GetCurrentWallpaper() (string, error) {
	currentWallpaperPath, err := GetCurrentWallpaperPathFromAPI()
	if err != nil {
		return "", fmt.Errorf("无法从API获取当前壁紙路径: %v", err)
	}
	if currentWallpaperPath == "" {
		return "", errors.New("从API获取到的壁纸路径为空, 可能没有设置壁纸")
	}
	log.Printf("通过API成功获取到当前壁纸路径: %s", currentWallpaperPath)


	// 定义保存的目标文件夹
	destDir := "image/raw_image"
	// 确保目标文件夹存在
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}

	// 打开源文件
	srcFile, err := os.Open(currentWallpaperPath)
	if err != nil {
		return "", fmt.Errorf("无法打开API提供的壁纸源文件 '%s': %v", currentWallpaperPath, err)
	}
	defer srcFile.Close()

	// 创建目标文件 (使用一个固定的、有意义的名字)
	destPath := filepath.Join(destDir, "backup_wallpaper"+filepath.Ext(currentWallpaperPath))
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
