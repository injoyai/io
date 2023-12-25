package io

import "time"

const (
	Ping = "ping"
	Pong = "pong"

	B   = 1         //1B
	KB  = 1024 * B  //1KB
	KB4 = 4 * KB    //4KB
	MB  = 1024 * KB //1MB
	GB  = 1024 * MB //1GB
	TB  = 1024 * GB //1TB
	PB  = 1024 * TB //1PB
	EB  = 1024 * PB //1EB

	DefaultBufferSize      = KB4              //默认buff大小,4KB
	DefaultChannelSize     = 100              //默认通道大小
	DefaultPort            = 10086            //默认端口
	DefaultPortStr         = ":10086"         //默认端口
	DefaultConnectTimeout  = time.Second * 2  //默认连接时间
	DefaultKeepAlive       = time.Minute * 10 //默认保持连接时间
	DefaultTimeoutInterval = time.Minute      //默认离线检查间隔
	DefaultResponseTimeout = time.Second * 10 //默认响应超时时间

)

const (
	B_TCP       = 0x00 // "TCP"
	B_UDP       = 0x01 // "UDP"
	B_HTTP      = 0x02 // "HTTP"
	B_Websocket = 0x03 // "Websocket"
	B_Memory    = 0x04 // "Memory"
	B_Serial    = 0x05 // "Serial"
	B_SSH       = 0x06 // "SSH"
	B_MQTT      = 0x07 // "MQTT"
)

const (
	TCP       = "tcp"
	UDP       = "udp"
	HTTP      = "http"
	Websocket = "websocket"
	Memory    = "memory"
	Serial    = "serial"
	SSH       = "ssh"
	MQTT      = "mqtt"
)

const (
	TypeConnect = 0x01 // "connect" 建立连接
	TypeWrite   = 0x02 // "write" 写入数据
	TypeClose   = 0x03 // "close" 关闭连接
)
