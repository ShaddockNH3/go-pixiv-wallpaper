package sys

import (
	"errors"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/image/bmp"
	"golang.org/x/sys/windows/registry"
)

type WallpaperStyle uint

func (wps WallpaperStyle) String() string {
	return wallpaperStyles[wps]
}

const (
	Fill                 WallpaperStyle = iota // 填充
	Fit                                        // 适应
	Stretch                                    // 拉伸
	Tile                                       // 平铺
	Center                                     // 居中
	Cross                                      // 跨区
	SPI_GETDESKWALLPAPER = 0x0073
)

var wallpaperStyles = map[WallpaperStyle]string{
	0: "填充",
	1: "适应",
	2: "拉伸",
	3: "平铺",
	4: "居中",
	5: "跨区"}

var (
	bgFile       string
	bgStyle      int
	sFile        string
	waitTime     int
	activeScreen bool
	passwd       bool
)

var (
	regist registry.Key
)

func init() {
	var err error
	regist, err = registry.OpenKey(registry.CURRENT_USER, `Control Panel\Desktop`, registry.ALL_ACCESS)
	if err != nil {
		// 在库中，我们通常不直接 log.Fatal，而是 panic 或返回错误
		// 这里为了简单，暂时保留
		log.Fatal(err)
	}

	libuser32 = MustLoadLibrary("user32.dll")
	libkernel32 = MustLoadLibrary("kernel32.dll")
	systemParametersInfo = MustGetProcAddress(libuser32, "SystemParametersInfoW")
	getVersion = MustGetProcAddress(libkernel32, "GetVersion")
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// http://blog.csdn.net/kfysck/article/details/8153264
// Check that the OS is Vista or later (Vista is v6.0).
func checkVersion() bool {
	version := GetVersion()
	major := version & 0xFF
	if major < 6 {
		return false
	}
	return true
}

// jpg转换为bmp
func ConvertedWallpaper(bgfile string) string {
	file, err := os.Open(bgfile)
	checkErr(err)
	defer file.Close()

	img, err := jpeg.Decode(file) //解码
	checkErr(err)

	bmpPath := os.Getenv("USERPROFILE") + `\Local Settings\Application Data\Microsoft\Wallpaper1.bmp`
	bmpfile, err := os.Create(bmpPath)
	checkErr(err)
	defer bmpfile.Close()

	err = bmp.Encode(bmpfile, img)
	checkErr(err)
	return bmpPath
}

func SetDesktopWallpaper(bgFile string, style WallpaperStyle) error {
	absBgFile, err := filepath.Abs(bgFile)
	if err != nil {
		return err
	}

	ext := filepath.Ext(absBgFile)
	if !checkVersion() && ext != ".bmp" {
		setRegistString("ConvertedWallpaper", absBgFile)
		absBgFile = ConvertedWallpaper(absBgFile)
	}
	setRegistString("Wallpaper", absBgFile)

	var bgTileWallpaper, bgWallpaperStyle string
	bgTileWallpaper = "0"
	switch style {
	case Fill:
		bgWallpaperStyle = "10"
	case Fit:
		bgWallpaperStyle = "6"
	case Stretch:
		bgWallpaperStyle = "2"
	case Tile:
		bgTileWallpaper = "1"
		bgWallpaperStyle = "0"
	case Center:
		bgWallpaperStyle = "0"
	case Cross:
		bgWallpaperStyle = "22"
	}

	setRegistString("WallpaperStyle", bgWallpaperStyle)
	setRegistString("TileWallpaper", bgTileWallpaper)

	ok := SystemParametersInfo(SPI_SETDESKWALLPAPER, 0, stringToPointer(absBgFile), SPIF_UPDATEINIFILE|SPIF_SENDWININICHANGE)
	if !ok {
		return errors.New("desktop background settings fail")
	}
	return nil
}

func setRegistString(name, value string) {
	err := regist.SetStringValue(name, value)
	checkErr(err)
}

func setScreenSaver(uiAction, uiParam uint32) {
	ok := SystemParametersInfo(uiAction, uiParam, nil, SPIF_UPDATEINIFILE|SPIF_SENDWININICHANGE)
	if !ok {
		log.Fatal("Screen saver Settings fail.")
	}
}

func getScreenSaver() bool {
	_, _, err := regist.GetStringValue("SCRNSAVE.EXE")
	return err == nil
}

func GetCurrentWallpaperPathFromAPI() (string, error) {
	const MAX_PATH = 300
	var buffer [MAX_PATH]uint16 // 使用 UTF-16 编码

	ok := SystemParametersInfo(SPI_GETDESKWALLPAPER, MAX_PATH, unsafe.Pointer(&buffer), 0)
	if !ok {
		return "", errors.New("调用 SPI_GETDESKWALLPAPER 失败")
	}

	return syscall.UTF16ToString(buffer[:]), nil
}
