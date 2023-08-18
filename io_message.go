package io

import (
	"fmt"
	"github.com/injoyai/base/bytes"
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

func (this *IMessage) Error() string {
	return this.Message.String()
}
