package io

import "time"

const (
	Ping = "ping"
	Pong = "pong"

	KB1               = 1 << 10  //1KB
	KB4               = 4 << 10  //4KB
	KB32              = 32 << 10 //4KB
	MB1               = 1 << 20  //1MB
	DefaultBufferSize = KB4

	DefaultPort            = 10086
	DefaultConnectTimeout  = time.Second * 2  //默认连接时间
	DefaultKeepAlive       = time.Minute * 10 //默认保持连接时间
	DefaultTimeoutInterval = time.Minute      //默认离线检查间隔
	DefaultResponseTimeout = time.Second * 10 //默认响应超时时间

)

const (
	B_TCP       = 0x00 //"TCP"
	B_UDP       = 0x01 // "UDP"
	B_HTTP      = 0x02 //"HTTP"
	B_Websocket = 0x03 // "Websocket"
	B_Memory    = 0x04 // "Memory"
	B_Serial    = 0x05 // "Serial"
	B_SSH       = 0x06 // "SSH"
	B_MQTT      = 0x07 // "MQTT"
)

const (
	TCP       = "TCP"
	UDP       = "UDP"
	HTTP      = "HTTP"
	Websocket = "Websocket"
	Memory    = "Memory"
	Serial    = "Serial"
	SSH       = "SSH"
	MQTT      = "MQTT"
)

const (
	TypeConnect = 0x01 // "connect"
	TypeWrite   = 0x02 // "write"
	TypeClose   = 0x03 // "close"
)
