package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func MoveAllFiles(srcDir, destDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("无法读取源目录 %s: %w", srcDir, err)
	}

	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return fmt.Errorf("无法创建目标目录 %s: %w", destDir, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())

		if entry.IsDir() {
			log.Printf("跳过子目录: %s", srcPath)
			continue
		}

		log.Printf("正在移动文件: %s -> %s", srcPath, destPath)
		if err := os.Rename(srcPath, destPath); err != nil {
			return fmt.Errorf("移动文件 %s 失败: %w", srcPath, err)
		}
	}

	return nil
}

func IsFilesEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		// 检查路径是否存在，或者是否是目录等错误
		return false, fmt.Errorf("无法读取目录 %s: %w", path, err)
	}

	// 检查条目列表的长度
	return len(entries) == 0, nil
}
