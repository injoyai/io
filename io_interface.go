package io

import (
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
	WriteCloser
}

// DialFunc 连接函数
type DialFunc func() (ReadWriteCloser, error)

// ListenFunc 监听函数
type ListenFunc func() (Listener, error)

// DealFunc 数据处理函数
type DealFunc func(msg Message)

// PrintFunc 打印函数
type PrintFunc func(msg Message, tag ...string)
