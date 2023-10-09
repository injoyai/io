package io

import (
	"github.com/injoyai/base/bytes"
)

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
