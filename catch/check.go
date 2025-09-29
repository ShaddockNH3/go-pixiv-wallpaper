package catch

import (
	"GopherPaws/config"
	"GopherPaws/resolution"
	"errors"
	"log"
	"time"
)

func Check() error {
	myconfig, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Println("加载config发生错误")
		return err
	}

	// 检查当日是否运行过了逻辑
	currentTime := time.Now()
	dateString := currentTime.Format("2006-01-02")
	if myconfig.Time.Now == dateString {
		return errors.New("Today Has Been Run")
	}

	// 然后是导入逻辑
	res, err := resolution.GetResolutionLogic()

	if err != nil {
		log.Println("获取res发生错误")
		return err
	}

	myconfig.Screen.Res = res
	myconfig.Time.Now = time.Now().Format("2006-01-02")

	err = config.SaveConfig("config/config.yaml", myconfig)
	if err != nil {
		log.Println("存储发生错误")
		return err
	}
	return nil
}
