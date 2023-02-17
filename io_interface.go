package io

import (
	"context"
	"time"
)

type BytesWriter interface {
	WriteBytes(p []byte) (int, error)
}

type TimeoutWriter interface {
	WriteWithTimeout(p []byte, timeout time.Duration) (int, error)
}

// Closed 是否已关闭
type Closed interface{ Closed() bool }

// Debugger 是否调试
type Debugger interface{ Debug(b ...bool) }

// Publisher 发布者
type Publisher interface {
	Publish(topic string, p []byte) error
}

// MessageReader 读取分包后的数据
type MessageReader interface {
	ReadMessage() ([]byte, error)
}

type MessageReadCloser interface {
	MessageReader
	Closer
}

type MessageReadWriteCloser interface {
	MessageReader
	Writer
	Closer
}

type _messageReadWriteCloser struct {
	MessageReadWriteCloser
}

func (this *_messageReadWriteCloser) Read(p []byte) (int, error) {
	return 0, nil
}

func NewMessageReadCloser(r MessageReadWriteCloser) ReadWriteCloser {
	return &_messageReadWriteCloser{MessageReadWriteCloser: r}
}

type Listener interface {
	Accept() (ReadWriteCloser, string, error)
	Close() error
	Addr() string
}

// DialFunc 连接函数
type DialFunc func() (ReadWriteCloser, error)

// ListenFunc 监听函数
type ListenFunc func() (Listener, error)

// DealFunc 数据处理函数
type DealFunc func(msg Message)

// PrintFunc 打印函数
type PrintFunc func(msg Message, tag ...string)

// CloseFunc 关闭函数
type CloseFunc func(ctx context.Context, msg Message)
