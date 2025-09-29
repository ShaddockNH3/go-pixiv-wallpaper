package main

import (
	"GopherPaws/catch"
	"GopherPaws/config"
	"GopherPaws/sys"
	"GopherPaws/utils"
	"log"
	"os"
)

func CreateDir(path string) {
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Printf("创建文件夹错误:%v\n", err)
		return
	}
}

func main() {
	CreateDir("image")
	CreateDir("image/raw_image")
	CreateDir("image/today_image")
	CreateDir("image/used_image")

	is_raw_empty, err := utils.IsFilesEmpty("image/raw_image")

	if err != nil {
		log.Printf("检查raw_image是否为空错误：%v", err)
	}

	if is_raw_empty {
		log.Printf("没有原始图片，将本地壁纸作为原始图片保存")
		_, err := sys.GetCurrentWallpaper()
		if err != nil {
			log.Printf("备份壁纸失败: %v\n", err)
		}
	}

	log.Println("开始执行每日检查...")
	err = catch.Check()
	if err != nil {
		if err.Error() == "Today Has Been Run" {
			log.Println("今天已经运行过啦，明天再见~")
			return
		}
		log.Fatalf("检查过程中发生错误: %v", err)
	}
	log.Println("检查完成，需要执行今天的壁纸寻找任务！")

	log.Println("将之前的壁纸移动到已使用文件夹")

	utils.MoveAllFiles("image/today_image", "image/used_image")

	log.Println("正在加载配置文件...")
	myconfig, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}
	log.Printf("成功加载配置，屏幕分辨率为: %dx%d", myconfig.Screen.Res.Width, myconfig.Screen.Res.Height)

	catch.RunWallpaperFinder(myconfig.Screen.Res.Width, myconfig.Screen.Res.Height, myconfig.Time.Now)

	is_today_empty, err := utils.IsFilesEmpty("image/today_image")
	if err != nil {
		log.Printf("判断今日图片为空的逻辑错误：%v", err)
	}

	var today_path string

	if is_today_empty {
		today_path = "image/raw_image"
	} else {
		today_path = "image/today_image"
	}

	err = sys.SetTodayWallpaper(today_path)
	if err != nil {
		log.Printf("设置壁纸错误:%v\n", err)
		return
	}

	log.Println("今日任务全部完成！")
}
