package io

import (
	"context"
	"github.com/injoyai/base/bytes"
	"time"
)

type Bytes = Message
type Message = bytes.Entity

type TimeoutWriter interface {
	WriteWithTimeout(p []byte, timeout time.Duration) (int, error)
}

type AnyWriter interface {
	WriteAny(any interface{}) (int, error)
}

type AnyWriterClosed interface {
	AnyWriter
	Closed
}

type Runner interface {
	Run() error
	Running() bool
}

// Closed 是否已关闭
type Closed interface {
	Closed() bool
}

type Closer2 interface {
	Closer
	Closed
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
	Closer
	Accept() (ReadWriteCloser, string, error)
	Addr() string
}

// ListenFunc 监听函数
type ListenFunc func() (Listener, error)

// DialFunc 连接函数
type DialFunc func(ctx context.Context) (ReadWriteCloser, string, error)

// OptionClient 客户端选项
type OptionClient func(c *Client)

// OptionServer 服务端选项
type OptionServer func(s *Server)

//=================================Key=================================

// ReadFunc 读取函数
type ReadFunc func(p []byte) (int, error)

func (this ReadFunc) Read(p []byte) (int, error) { return this(p) }

// WriteFunc 写入函数
type WriteFunc func(p []byte) (int, error)

func (this WriteFunc) Write(p []byte) (int, error) { return this(p) }

// CloseFunc 关闭函数
type CloseFunc func() error

func (this CloseFunc) Close() error { return this() }

type Key string

func (this *Key) GetKey() string { return string(*this) }

func (this *Key) SetKey(key string) { *this = Key(key) }
