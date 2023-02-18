package io

import (
	"fmt"
	"github.com/injoyai/base/bytes"
)

const (
	TagRead  = "接收"
	TagWrite = "发送"
	TagErr   = "错误"
	TagInfo  = "信息"
	Ping     = "ping"
	Pong     = "pong"
)

func NewMessageFormat(format string, v ...interface{}) Message {
	return NewMessage(fmt.Sprintf(format, v...))
}

func NewMessage(s string) Message {
	return []byte(s)
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
