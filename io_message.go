package io

import (
	"fmt"
	"github.com/injoyai/base/bytes"
)

const (
	TagRead  = "接收"
	TagWrite = "发送"
	TagClose = "关闭"
	TagDial  = "连接"
	TagErr   = "错误"
	Ping     = "ping"
	Pong     = "pong"
)

func NewMessageFormat(format string, v ...interface{}) Message {
	return NewMessageString(fmt.Sprintf(format, v...))
}

func NewMessageString(s string) Message {
	return NewMessage([]byte(s))
}

func NewMessage(bs []byte) Message {
	return bs
}

type Message = bytes.Entity

func NewClientMessage(c *Client, p []byte) *ClientMessage {
	return &ClientMessage{
		Client:  c,
		Message: p,
	}
}

type ClientMessage struct {
	*Client
	Message
}
