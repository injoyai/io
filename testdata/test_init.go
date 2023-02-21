package testdata

import "github.com/injoyai/logs"

func init() {
	//设置只打印到终端
	logs.SetWriter(logs.Stdout)
}
