package main


import "geekbang/download"
func main() {

	// 得到json文件
	download.FindLink()
	// 创建文件夹
	download.CreateFolder()
	download.DownImages2()
}

