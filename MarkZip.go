package main

/**
处理对文件进行加密解密处理
**/

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

//加密文件存储路径
func getMPath(filePath string) string {
	p, n := getPathAndFile(filePath)
	return p + "/m_" + n
}

//解密文件存储路径
func getUPath(filePath string) string {
	p, n := getPathAndFile(filePath)
	return p + "/u_" + n
}

//对文件进行加密处理
func MarkFile(filePath string) {
	MarkFileTo(filePath, "", false)
}

//对文件进行加密处理
func MarkFileTo(filePath string, targetPath string, isDelSrc bool) {
	zipm := ZipModel{}
	zipm.init()
	//读取文件
	fr, _ := os.Open(filePath)
	defer fr.Close()
	//输出文件
	p, n := getPathAndFile(filePath)
	if targetPath == "" {
		targetPath = p
	}
	//文件名加上父路径
	var parentPath = getParentDirectory(filePath)
	var outpath = targetPath + "/m_" + parentPath + "_" + n

	//1.初始化,判断读取的文件，如果读取的文件是zip文件，则直接返回
	zipm.ReLoadByFile(fr, false)
	if zipm.IsZip {
		return
	}

	//2.判断输出的文件，如果输出的文件已经存在，并且长度和md5都一致，则不处理
	outzipm := ZipModel{}
	outzipm.init()
	fhw, _ := os.Open(outpath)
	defer fhw.Close()

	outzipm.ReLoadByFile(fhw, false)
	logDebug(outzipm)
	if outzipm.Slen == zipm.Slen && outzipm.Smd532 == zipm.Smd532 {
		//保存到数据库中
		zipm.Zipstate = 1
		saveMovie(&zipm)
		logDebug(fhw.Name() + " has")
		return
	}

	//最大的文件不能超过10G，如果超过10G请先分段处理
	if zipm.Slen > 1024*1024*1000*10 {
		fmt.Println("file to big")
		return
	}

	//输出
	fw, _ := os.Create(outpath)
	defer fw.Close()
	//
	//转换成json
	zipd, _ := zipm.toJson()
	//头部固话加密5
	for i := 0; i < len(zipd); i++ {
		zipd[i] = zipd[i] + 5
	}

	//头部4位用于计算头部的长度 int
	var bufh = bytes.NewBuffer([]byte{})
	lenh := int32(len(zipd))
	binary.Write(bufh, binary.BigEndian, lenh)
	fw.Write(bufh.Bytes())
	fw.Write(zipd)
	//每次输出32KB
	var temp = make([]byte, 1024*64)
	for {
		rn, err := fr.Read(temp)
		if err != nil || rn <= 0 {
			break
		}
		for i := 0; i < rn; i++ {
			temp[i] = temp[i] + zipm.Byteadd
		}
		fw.Write(temp[:rn])
	}
	fr.Close()

	//zipm.save()
	zipm.Zipstate = 1
	saveMovie(&zipm)

	if isDelSrc {
		err := os.Remove(filePath) //删除文件test.txt
		if err != nil {
			//如果删除失败则输出 file remove Error!
			fmt.Println("file remove Error!")
			//输出错误详细信息
			fmt.Printf("%s", err)
		} else {
			//如果删除成功则输出 file remove OK!
			fmt.Println(filePath + "file remove OK!")
		}
	}
}

//对文件进行解密处理
func UnMarkFile(filePath string, isDelSrc bool) (string, error) {
	zipm := ZipModel{}
	zipm.init()
	//读取文件
	fr, _ := os.Open(filePath)
	defer fr.Close()

	//1.初始化,如果源来文件不存在，或者源来文件不是zip格式的，则直接返回
	zipm.ReLoadByFile(fr, false)
	if !zipm.IsZip {
		return "", errors.New("file is not zipfile")
	}

	//2.判断输出文件是否存在，如果已存在，并且长度相同，不重复处理
	outzipm := ZipModel{}
	outzipm.init()
	fhw, _ := os.Open(getUPath(filePath))
	defer fhw.Close()
	err := outzipm.ReLoadByFile(fhw, false)

	if outzipm.Slen == zipm.Slen && outzipm.Smd532 == zipm.Smd532 {
		logDebug(fhw.Name() + " has")
		return fhw.Name(), err
	}

	//输出文件
	fw, err := os.Create(getUPath(filePath))
	defer fw.Close()
	fr.Seek(int64(zipm.Hreadlen), os.SEEK_SET)
	//每次输出32KB
	var temp = make([]byte, 1024*32)
	for {
		rn, err := fr.Read(temp)
		if err != nil || rn <= 0 {
			break
		}
		for i := 0; i < rn; i++ {
			temp[i] = temp[i] - zipm.Byteadd
		}
		fw.Write(temp[:rn])
	}
	fr.Close()

	//删除原文件
	if isDelSrc {
		err := os.Remove(filePath)
		if err != nil {
			//如果删除失败则输出 file remove Error!
			//fmt.Println("file remove Error!")
			//输出错误详细信息
			//fmt.Printf("%s", err)
		} else {
			//如果删除成功则输出 file remove OK!
			//fmt.Println(filePath + "file remove OK!")
		}
	}

	return fw.Name(), err
}

//删除重复的文件
func CleanRepeatFile(basePath string, isNextPath bool) {
	var files = make([]string, 0, 10)

	if isNextPath {
		fs, _ := WalkDir(basePath, "", 0)
		files = fs
	} else {
		fs, _ := ListDir(basePath, "", 0)
		files = fs
	}

	var m1 map[string]ZipModel
	m1 = make(map[string]ZipModel)
	for _, f := range files {
		zipm := ZipModel{}
		zipm.init()
		//读取文件
		fr, _ := os.Open(f)
		defer fr.Close()
		zipm.ReLoadByFile(fr, false)
		fr.Close()

		//已经压缩的不算
		if zipm.IsZip {
			continue
		}
		// 查找键值是否存在
		if v, ok := m1[zipm.Smd532]; ok {
			if v.Slen != zipm.Slen {
				continue
			}
			err := os.Remove(f) //删除文件test.txt
			if err != nil {
				//如果删除失败则输出 file remove Error!
				fmt.Println("file remove Error!")
				//输出错误详细信息
				fmt.Printf("%s", err)
			} else {
				//如果删除成功则输出 file remove OK!
				fmt.Println(f + "file remove OK!")
			}
		} else {
			m1[zipm.Smd532] = zipm
			//fmt.Println(v)
		}
		//fmt.Println(f)
	}
}

//检查超出10G大小的影片，并输出
func checkMaxLen(basePath string, isNextPath bool) {
	var files = make([]string, 0, 10)

	if isNextPath {
		fs, _ := WalkDir(basePath, "", 0)
		files = fs
	} else {
		fs, _ := ListDir(basePath, "", 0)
		files = fs
	}

	for _, f := range files {
		zipm := ZipModel{}
		zipm.init()
		//读取文件
		fr, _ := os.Open(f)
		defer fr.Close()

		zipm.ReLoadByFile(fr, false)
		fr.Close()
		fmt.Println(f)
		if zipm.Slen > 1024*1024*1024*10 {
			fmt.Println("-----TO MAX-----:" + zipm.FullFileName)
		}
	}
}

//对文件进行压缩处理，压缩到指定目录
func MarkFilePath(srcPath, toPath string, maxMark int, isDelSrc bool) {
	files, _ := WalkDir(srcPath, "", 0)
	var count = 0
	for i, f := range files {
		//处理过的不再处理
		zipm, err := getMovieByFullFileName(f)
		if err == nil && zipm.Zipstate == 1 {
			continue
		}

		//只处理视频类文件
		var Upperfn = strings.ToUpper(f)
		movietype := []string{".MP4", ".AVI", ".MKV", ".WMV", ".RMVB", ".FLV"}
		var ismovie = false
		for _, mot := range movietype {
			if strings.Index(Upperfn, mot) > 0 {
				ismovie = true
			}
		}
		if !ismovie {
			continue
		}

		//只处理特殊的文件
		if strings.Index(Upperfn, "[FHD-1080P]") < 0 {
			continue
		}

		fmt.Println(strconv.Itoa(i) + ":" + strconv.Itoa(count) + ":" + f)

		MarkFileTo(f, toPath, isDelSrc)
		if count >= maxMark {
			break
		}

		count = count + 1
	}
}
