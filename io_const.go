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

	DefaultConnectTimeout  = time.Second * 2  //默认连接时间
	DefaultKeepAlive       = time.Minute * 10 //默认保持连接时间
	DefaultTimeoutInterval = time.Minute      //默认离线检查间隔
	DefaultResponseTimeout = time.Second * 10 //默认响应超时时间

)

const (
	TCP       = 0x00 //"TCP"
	UDP       = 0x01 // "UDP"
	HTTP      = 0x02 //"HTTP"
	Websocket = 0x03 // "Websocket"
	Memory    = 0x04 // "Memory"
	Serial    = 0x05 // "Serial"
	SSH       = 0x06 // "SSH"
	MQTT      = 0x07 // "MQTT"
)

const (
	TypeConnect = 0x01 // "connect"
	TypeWrite   = 0x02 // "write"
	TypeClose   = 0x03 // "close"
)
