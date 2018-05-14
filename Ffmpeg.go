// Ffmpeg
package main

import (
	//"fmt"
	"os"
	"os/exec"
)

//对视频文件进行截屏
//filePath:截取视频的文件
//ss:是截取视频的第几秒
//target:是输出视频的路径,如果不传，则输出到源目录下
func CatMovieImage(filePath string, ss string, target string) bool {
	//path, fn := getPathAndFile(filePath)
	var imgp = filePath + ss + ".jpg"
	if len(target) > 2 {
		imgp = target
	}
	cmd := exec.Command("ffmpeg", "-ss", ss, "-i", filePath, "-y", "-s", "800x600", "-f", "image2", imgp)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		//fmt.Println("Error: ", err)
	}
	_, err := os.Stat(imgp)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}
