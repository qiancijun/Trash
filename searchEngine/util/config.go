package util

import (
	"path"
	"runtime"
)

var (
	RootPath string
)

func init() {
	RootPath = path.Dir(GetCurrentPath() + "..") + "/"
}

func GetCurrentPath() string {
	_, filename, _, _ := runtime.Caller(1) // 1 表示当前函数，2 表示调用本函数的函数
	return path.Dir(filename)
}