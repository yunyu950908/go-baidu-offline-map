package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	P4L string  = "../node_modules/proj4leaflet/"
	R   float64 = 6378137 // 地球半径
	// qt=vtile, scaler=2, udt=20190122 为变量，根据最新接口自行修改
	BASE_URL string = "http://online1.map.bdimg.com/tile/?qt=vtile&styles=pl&scaler=2&udt=20190122&z="
)

// 默认为当前时间的时间戳
var targetDir string = time.Now().Format("20060102150405")

// 阻塞队列
var waitgroup sync.WaitGroup

func main() {
	fmt.Println("****************************")

	var mins, maxs, confirm, fls, targetPath string

	fmt.Printf("输入最小、大层级（半角逗号隔开）：")
	fmt.Scanln(&fls)

	fmt.Printf("输入最小经、纬度（半角逗号隔开）：")
	fmt.Scanln(&mins)

	fmt.Printf("输入最大经、纬度（半角逗号隔开）：")
	fmt.Scanln(&maxs)

	fmt.Printf("输入存储瓦片的文件夹名称：")
	fmt.Scanln(&targetDir)

	if fls == "" || mins == "" || maxs == "" {
		fmt.Println("输入值为空！")
		return
	}

	targetPath = "./" + targetDir + "/tiles"

	// 切割字符串
	mina := strings.Split(mins, ",")
	maxa := strings.Split(maxs, ",")
	fla := strings.Split(fls, ",")

	minLng, _ := strconv.ParseFloat(mina[0], 64)
	minLat, _ := strconv.ParseFloat(mina[1], 64)
	maxLng, _ := strconv.ParseFloat(maxa[0], 64)
	maxLat, _ := strconv.ParseFloat(maxa[1], 64)

	startZ, _ := strconv.Atoi(fla[0])
	endZ, _ := strconv.Atoi(fla[1])

	// 取得地图中心坐标经纬度
	lngCen := (minLng + maxLng) / 2.0
	latCen := (minLat + maxLat) / 2.0

	fmt.Printf("数据输入正确，是否开始下载？（Y/n）：")
	fmt.Scanln(&confirm)

	if confirm == "n" {
		fmt.Println("程序退出！")
		return
	}

	fmt.Println("---------------------下载开始---------------------")

	// 开始执行时间
	startTime := time.Now()

	GetAllFloor(minLng, maxLng, minLat, maxLat, startZ, endZ, targetPath)

	// 计算执行耗时
	allTime := time.Since(startTime)

	fmt.Println(allTime)

	// 生成地图索引文件
	MkIndex(lngCen, latCen, fla[0], fla[1])

	fmt.Println("---------------------下载完成---------------------")

	fmt.Println("打开 index.html 即可预览离线地图，程序退出！")
}

// 判断所给路径文件/文件夹是否存在
func isExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// 获得所有层级瓦片图
func GetAllFloor(minLng float64, maxLng float64, minLat float64, maxLat float64, startZ int, endZ int, targetPath string) {
	// 所有层级文件夹所在的文件夹
	if !isExist(targetPath) {
		os.MkdirAll(targetPath, os.ModePerm)
	}

	for z := startZ; z <= endZ; z++ {
		waitgroup.Add(1)
		go GetOneFloor(minLng, maxLng, minLat, maxLat, z, targetPath)
		fmt.Println(strconv.Itoa(z))
	}

	waitgroup.Wait()
}

// 获得一个层级瓦片图
func GetOneFloor(minLng float64, maxLng float64, minLat float64, maxLat float64, z int, targetPath string) {
	url := BASE_URL + strconv.Itoa(z)

	minX, maxX, minY, maxY := GetBound(minLng, maxLng, minLat, maxLat, z)

	var urlPath, picPath string
	var dir string

	fdir := targetPath + "/" + strconv.Itoa(z) + "/"

	// 层级文件夹
	if !isExist(fdir) {
		os.Mkdir(fdir, os.ModePerm)
	}

	for i := minX; i <= maxX; i++ {
		func(i int, minY int, maxY int) {
			dir = fdir + strconv.Itoa(i)
			// 瓦片文件夹
			if !isExist(dir) {
				os.Mkdir(dir, os.ModePerm)
			}
			for j := minY; j <= maxY; j++ {
				// pic 文件不存在 ==> 继续
				picPath = dir + "/" + strconv.Itoa(j) + ".png"
				if !isExist(picPath) {
					urlPath = url + "&x=" + strconv.Itoa(i) + "&y=" + strconv.Itoa(j)
					resp, _ := http.Get(urlPath)
					body, _ := ioutil.ReadAll(resp.Body)
					out, _ := os.Create(picPath)
					io.Copy(out, bytes.NewReader(body))
				}
			}
		}(i, minY, maxY)
	}

	waitgroup.Done()
}

// 根据经纬度和层级转换瓦片图范围
func GetBound(minLng float64, maxLng float64, minLat float64, maxLat float64, z int) (minX int, maxX int, minY int, maxY int) {
	minX = int(math.Floor(math.Pow(2.0, float64(z-26)) * (math.Pi * minLng * R / 180.0)))
	maxX = int(math.Floor(math.Pow(2.0, float64(z-26)) * (math.Pi * maxLng * R / 180.0)))

	minY = int(math.Floor(math.Pow(2.0, float64(z-26)) * R * math.Log(math.Tan(math.Pi*minLat/180.0)+1.0/math.Cos(math.Pi*minLat/180.0))))
	maxY = int(math.Floor(math.Pow(2.0, float64(z-26)) * R * math.Log(math.Tan(math.Pi*maxLat/180.0)+1.0/math.Cos(math.Pi*maxLat/180.0))))

	return
}

// 生成地图索引文件
func MkIndex(lngCen, latCen float64, startZ string, endZ string) {
	fname := "./" + targetDir + "/index.html"
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModeAppend|os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	content := `
<!DOCTYPE html>
<html>
    <head>
        <meta charset="utf-8">
        <title>地图预览</title>
        <link rel="stylesheet" href="` + P4L + `lib/leaflet/leaflet.css" />
        <style>
            #map {height: 800px;}
        </style>
    </head>
    <body>
        <div id="map"></div>
        
        <script src="` + P4L + `lib/leaflet/leaflet.js"></script>
        <script src="` + P4L + `lib/proj4-compressed.js"></script>
        <script src="` + P4L + `src/proj4leaflet.js"></script>
        <script>
            var center = {
                lng: "` + strconv.FormatFloat(lngCen, 'f', -1, 64) + `",
                lat: "` + strconv.FormatFloat(latCen, 'f', -1, 64) + `"
            }
            
            // 百度坐标转换
            var crs = new L.Proj.CRS(
                'EPSG:3395',
                '+proj=merc +lon_0=0 +k=1 +x_0=0 +y_0=0 +datum=WGS84 +units=m +no_defs',
                {
                    resolutions: function () {
                        level = 19
                        var res = [];
                        res[0] = Math.pow(2, 18);
                        for (var i = 1; i < level; i++) {
                            res[i] = Math.pow(2, (18 - i))
                        }
                        return res;
                    }(),
                    origin: [0, 0],
                    bounds: L.bounds([20037508.342789244, 0], [0, 20037508.342789244])
                }
            );
            var map = L.map('map', { crs: crs });
            
            L.tileLayer('./tiles/{z}/{x}/{y}.png', {
                maxZoom: ` + endZ + `,
                minZoom: ` + startZ + `,
                subdomains: [0,1,2],
                tms: true
            }).addTo(map);
            
            new L.marker([center.lat, center.lng]).addTo(map);
            
            map.setView([center.lat, center.lng], ` + startZ + `);
        </script>
    </body>
</html>`

	f.WriteString(content)
	f.Close()
}
