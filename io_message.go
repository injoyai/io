package io

import (
	"fmt"
	"github.com/injoyai/base/bytes"
	"time"
)

const (
	TagRead  = "接收"
	TagWrite = "发送"
	TagErr   = "错误"
	TagInfo  = "信息"
	Ping     = "ping"
	Pong     = "pong"

	DefaultKeepAlive       = time.Minute * 10 //默认保持连接时间
	DefaultTimeoutInterval = time.Minute      //默认离线检查间隔
	DefaultResponseTimeout = time.Second * 10 //默认响应超时时间
)

func NewMessageFormat(format string, v ...interface{}) Message {
	return Message(fmt.Sprintf(format, v...))
}

type Message = bytes.Entity

func NewIMessage(c *Client, p []byte) *IMessage {
	return &IMessage{
		Client:  c,
		Message: p,
	}
}

type IMessage struct {
	*Client
	Message
}
