package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"crypto/md5"
	"encoding/binary"

	"io"
	"os"
)

//获取文件的MD5
func ComputeMd5(file *os.File) ([]byte, error) {
	var result []byte
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}
	file.Seek(0, os.SEEK_SET)
	return hash.Sum(result), nil
}

func ComputeMd5By32(file *os.File) ([]byte, error) {
	var result []byte

	//获取文件大小
	buf_len, _ := file.Seek(0, os.SEEK_END)
	file.Seek(0, os.SEEK_SET)
	//只取10块内容做MD5
	var clean = buf_len / 10
	logDebug(clean)
	hash := md5.New()
	var temp = make([]byte, 1024)
	var count = 0
	for {
		rn, err := file.Read(temp)
		if err != nil || rn <= 0 {
			break
		}
		file.Seek(clean, os.SEEK_CUR)
		hash.Write(temp)
		count = count + 1
		if count > 10 {
			//break
		}
	}

	len_b := make([]byte, 8)
	binary.BigEndian.PutUint64(len_b, uint64(buf_len))
	hash.Write(len_b)
	file.Seek(0, os.SEEK_SET)
	return hash.Sum(result), nil
}

//上级str
func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

//获取上级路径
func getParentDirectory(dirctory string) string {
	dirctory = strings.Replace(dirctory, "\\", "/", -1)
	var sp = strings.LastIndex(dirctory, "/")
	ps := strings.Split(dirctory, "/")
	if sp == len(dirctory) {
		return ps[len(ps)-3]
	}
	return ps[len(ps)-2]
}

//获取当前路径
func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

//获取文件路径和文件名
func getPathAndFile(filepath string) (string, string) {
	var fp = strings.Replace(filepath, "\\", "/", -1)

	var sp = strings.LastIndex(fp, "/")
	//fmt.Print(sp)
	return fp[0:sp], fp[sp+1:]
	//return substr(filepath, 0, sp), substr(filepath, sp+1, len(filepath)-sp)
}

//获取指定目录下的所有文件，不进入下一级目录搜索，可以匹配后缀过滤。
func ListDir(dirPth string, suffix string, minlen int64) (files []string, err error) {
	files = make([]string, 0, 10)

	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}
		//小于指定大小的忽略
		if minlen > 0 && minlen < fi.Size() {
			continue
		}
		//后缀名匹配

		if len(suffix) > 0 {
			if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
				files = append(files, dirPth+PthSep+fi.Name())
			}
		} else {
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}

	return files, nil
}

//获取指定目录及所有子目录下的所有文件，可以匹配后缀过滤。
func WalkDir(dirPth, suffix string, minlen int64) (files []string, err error) {
	files = make([]string, 0, 30)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error { //遍历目录
		//if err != nil { //忽略错误
		// return err
		//}

		if fi.IsDir() { // 忽略目录
			return nil
		}
		//小于指定大小的忽略
		if minlen > 0 && minlen < fi.Size() {
			return nil
		}
		//后缀名称匹配
		if len(suffix) > 0 {
			if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
				files = append(files, filename)
			}
		} else {
			files = append(files, filename)
		}
		return nil
	})

	return files, err
}

// 获取本机的MAC地址
func getMac() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Error : " + err.Error())
	}
	for _, inter := range interfaces {
		mac := inter.HardwareAddr //获取本机MAC地址
		fmt.Println("MAC = ", mac)
		return mac.String()
	}
	return ""
}

// 简单直接的GET请求
func httpGet(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
