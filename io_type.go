package io

import (
	"bufio"
	"context"
	"github.com/injoyai/base/bytes"
	"time"
)

type Message = bytes.Entity

type Model struct {
	Type string      `json:"type"`           //请求类型,例如测试连接ping,写入数据write... 推荐请求和响应通过code区分
	Code int         `json:"code,omitempty"` //请求结果,推荐 请求:0(或null)  响应: 200成功,500失败... 同http好记一点
	UID  string      `json:"uid,omitempty"`  //消息的唯一ID,例如UUID
	Data interface{} `json:"data,omitempty"` //请求响应的数据
	Msg  string      `json:"msg,omitempty"`  //消息
}

type TimeoutWriter interface {
	WriteWithTimeout(p []byte, timeout time.Duration) (int, error)
}

type Runner interface {
	Run() error
	Running() bool
}

// Closed 是否已关闭
type Closed interface {
	Closed() bool
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

// DialFunc 连接函数
type DialFunc func() (ReadWriteCloser, string, error)

// ReadFunc 读取函数
type ReadFunc func(buf *bufio.Reader) ([]byte, error)

// WriteFunc 写入函数
type WriteFunc func(p []byte) ([]byte, error)

// DealFunc 数据处理函数
type DealFunc func(msg Message) error

// CloseFunc 关闭函数
type CloseFunc func(ctx context.Context, msg Message)

// WriteDeadline 写入超时时间,例如tcp关闭
type WriteDeadline func(t time.Time) error

// OptionClient 客户端选项
type OptionClient func(c *Client)

// OptionServer 服务端选项
type OptionServer func(s *Server)

//=================================Key=================================

type Key struct{ key string }

func (this *Key) GetKey() string { return this.key }

func (this *Key) SetKey(key string) { this.key = key }

//=================================Debugger=================================

type Debugger bool

func (this *Debugger) Debug(b ...bool) {
	*this = Debugger(len(b) == 0 || b[0])
}
