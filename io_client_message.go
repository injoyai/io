package io

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
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

func NewMessageStringf(format string, v ...interface{}) Message {
	return NewMessage([]byte(fmt.Sprintf(format, v...)))
}

func NewMessageErr(err error) Message {
	if err != nil {
		return NewMessageString(err.Error())
	}
	return Message{}
}

func NewMessage(bs []byte) Message {
	return bs
}

type Message []byte

func (this Message) Error() string {
	return this.String()
}

func (this Message) String() string {
	return string(this.Bytes())
}

func (this Message) HEX() string {
	return strings.ToUpper(hex.EncodeToString(this.Bytes()))
}

func (this Message) ASCII() string {
	return string(this.Bytes())
}

func (this Message) Bytes() []byte {
	return this
}

func (this Message) Base64() string {
	return base64.StdEncoding.EncodeToString(this.Bytes())
}

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
