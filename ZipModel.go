package main

/**
1、golang的命名需要使用驼峰命名法，且不能出现下划线
2、golang中根据首字母的大小写来确定可以访问的权限。无论是方法名、常量、变量名还是结构体的名称，如果首字母大写，则可以被其他的包访问；如果首字母小写，则只能在本包中使用
可以简单的理解成，首字母大写是公有的，首字母小写是私有的
**/

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	//"github.com/ablegao/orm"
	//"github.com/go-sql-driver/mysql"
	"github.com/tidwall/gjson"
)

type ZipModel struct {
	//原文件名称（文件全称，不包括路径）
	Sname string `field:"Sname"`

	//文件的全路径
	FullFileName string `field:"FullFileName"`

	//原文件的md5
	Smd5 string `field:"Smd5"`

	//原文件的md5,32分之1
	Smd532 string `field:"smd532" auto:"false" index:"pk"`

	//原文件长度，byte
	Slen int64 `field:"Slen"`
	//原文件的创建时间
	Sctime int64 `field:"Sctime"`

	//创建时间
	CreateTime int64 `field:"CreateTime"`
	//版本号
	Version int `field:"Version"`

	//定义头部的长度，小于这个长度的，自动补全此长度，默认长度定义为10KB的长度。
	Hreadlen int `field:"Hreadlen"`
	//是否已经进行了处理
	IsZip bool
	//定义byte加多少
	Byteadd byte `field:"Byteadd"`

	//处理状态，0未处理，1已处理
	Zipstate int16 `field:"Zipstate"`
}

//数据库表名称
func (self *ZipModel) GetTableName() string {
	return "t_mv_movie"
}

//对结构体进行初始
func (item *ZipModel) init() {
	item.Hreadlen = 1024 * 10
	//创建时间，用整数表示
	item.CreateTime = time.Now().Unix()

	item.IsZip = false

	item.Byteadd = 100
	//
	//orm.NewDatabase("default", "mysql", "root:888@tcp(localhost:3333)/map")
}

//转换成json对象
func (item *ZipModel) toJson() ([]byte, error) {
	return json.Marshal(item)
}

//通过byte 创建对象
func (item *ZipModel) MarkZipModel(b []byte) error {
	var str = string(b)
	//logDebug(str)
	//str = substr(str, 0, strings.Index(str, "}")+1)
	var data []byte = []byte(str)

	if len(str) < 10 {
		return errors.New("no has zip heared")
	}

	//err := gjson.Unmarshal(data, item)
	err := gjson.Unmarshal(data, item)
	//logDebug(gjson.Get(str, "Sname"))
	//logDebug(err)
	return err
}

//通过文件重新初始化对象
func (zipm *ZipModel) ReLoadByFile(file *os.File, ismd5 bool) error {
	//头部4位用于计算头部的长度 int
	var bfhlen = make([]byte, 4)
	rn1, err1 := file.Read(bfhlen)
	if err1 != nil || rn1 < 4 {
		logDebug(err1)
		return err1
	}
	var hlen = int(binary.BigEndian.Uint32(bfhlen))
	if hlen > 1024*10 {
		hlen = 1024 * 10
	}
	if hlen < 2 {
		hlen = 2
	}
	//取头部
	var htemp = make([]byte, hlen)
	rn, err := file.Read(htemp)
	if err != nil {
		logDebug(err)
		return err
	}

	//头部处理
	//头部固化解密-5
	for i := 0; i < len(htemp); i++ {
		htemp[i] = htemp[i] - 5
	}

	file.Seek(0, os.SEEK_SET)
	//获取文件大小
	buf_len, _ := file.Seek(0, os.SEEK_END)
	file.Seek(0, os.SEEK_SET)

	//如果已经处理过了，直接返回头部描述
	if rn >= hlen {
		err := zipm.MarkZipModel(htemp)
		if err == nil {
			zipm.Hreadlen = len(htemp) + 4
			//判断文件大小符合要求的才算，避免部分文件输出不完整的不算
			if buf_len == zipm.Slen+int64(zipm.Hreadlen) {
				zipm.IsZip = true
				return err
			} else {
				fmt.Println("error len")
			}
		} else {
			logDebug(err)
			//return err
		}
	}

	zipm.Slen = buf_len
	//文件的创建时间
	fi, _ := file.Stat()
	zipm.Sctime = fi.ModTime().Unix()
	//文件的名称
	zipm.Sname = fi.Name()
	zipm.FullFileName = file.Name()
	//文件的MD5
	md532, _ := ComputeMd5By32(file)
	zipm.Smd532 = fmt.Sprintf("%x", md532)
	//fmt.Println(zipm)
	if ismd5 {
		md5, _ := ComputeMd5(file)
		zipm.Smd5 = fmt.Sprintf("%x", md5)
	}
	return err
}
