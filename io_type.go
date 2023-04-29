package io

import (
	"bufio"
	"context"
	"time"
)

type TimeoutWriter interface {
	WriteWithTimeout(p []byte, timeout time.Duration) (int, error)
}

type Runner interface {
	Run() error
}

// Closed 是否已关闭
type Closed interface {
	Closed() bool
}

// Debugger 是否调试
type Debugger interface {
	Debug(b ...bool)
}

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

type Listener interface {
	Accept() (ReadWriteCloser, string, error)
	Close() error
	Addr() string
}

// ListenFunc 监听函数
type ListenFunc func() (Listener, error)

// PrintFunc 打印函数
type PrintFunc func(msg Message, tag ...string)

// DialFunc 连接函数
type DialFunc func() (ReadWriteCloser, error)

// ReadFunc 读取函数
type ReadFunc func(buf *bufio.Reader) ([]byte, error)

// WriteFunc 写入函数
type WriteFunc func(p []byte) ([]byte, error)

// DealFunc 数据处理函数
type DealFunc func(msg Message)

// CloseFunc 关闭函数
type CloseFunc func(ctx context.Context, msg Message)

// WriteDeadline 写入超时时间,例如tcp关闭
type WriteDeadline func(t time.Time) error

type OptionClient func(ctx context.Context, c *Client)

type OptionServer func(s *Server)
