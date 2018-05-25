// easyZip project main.go
package main

import (
	"fmt"
	"time"
)

//	"time"
//	"bufio"

//"io/ioutil"

func main() {
	currtime := time.Now().Unix()
	//files, _ := WalkDir("e:/", "mp4", 0)
	//fmt.Println(len(files))
	//MarkFile("d:/temp/tomcat7_sfz-stdout.2017-11-06.log")
	//MarkFile("E:\\video-a\\1pondo-101916_408-FHD\\101916_408-1pon-1080p.mp4")
	//UnMarkFile("E:\\BaiduNetdiskDownload\\m_1025 -525_102517-525-carib-1080p.mp4", false)
	//UnMarkFile("E:\\mark\\m_Tokyo-Hot-RED-187-FHD_RED-187_sp_fhd_1.mp4")
	//CleanRepeatFile("E:/video", true)
	MarkFilePath("E:/video", "E:/markOther", 1000, false)
	//checkMaxLen("E:/video", true)
	//iscat := CatMovieImage("E:/video/[2D-FHD]SM3D2DBD-25/SM3D2DBD-25_22.mp4", "240", "e:/c222.jpg")
	//fmt.Println(iscat)

	fmt.Println(getMac())
	fmt.Println(time.Now().Unix() - currtime)

}
