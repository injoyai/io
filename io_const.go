package io

import "time"

const (
	TagRead  = "接收"
	TagWrite = "发送"
	TagErr   = "错误"
	TagInfo  = "信息"
	Ping     = "ping"
	Pong     = "pong"

	KB1               = 1 << 10 //1KB
	KB4               = 4 << 10 //4KB
	MB1               = 1 << 20 //1MB
	DefaultBufferSize = KB4

	DefaultKeepAlive       = time.Minute * 10 //默认保持连接时间
	DefaultTimeoutInterval = time.Minute      //默认离线检查间隔
	DefaultResponseTimeout = time.Second * 10 //默认响应超时时间
)
