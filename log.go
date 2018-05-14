package main

/**
实现简单的日志输出
分为3种输出，调试输出，信息输出，错误输出3类
**/

import (
	"fmt"
)

func logInfo(a ...interface{}) {
	fmt.Println(a)
}

func logDebug(a ...interface{}) {
	//fmt.Println(a)
}

func logError(a ...interface{}) {
	fmt.Println(a)
}
