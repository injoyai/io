package io

import (
	"bufio"
	"context"
	"github.com/injoyai/base/bytes"
	"io"
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

//=================

// MustChan chan []byte 实现 io.Writer,必须等到写入成功为止
type MustChan chan []byte

func (this MustChan) Write(p []byte) (int, error) {
	this <- p
	return len(p), nil
}

// TryChan chan []byte 实现 io.Writer,尝试写入,不管是否成功
type TryChan chan []byte

func (this TryChan) Write(p []byte) (int, error) {
	select {
	case this <- p:
		return len(p), nil
	default:
		return 0, nil
	}
}

//====================

// Count 统计写入字节数量
type Count struct {
	io.Writer
	count int64
}

func (this *Count) Count() int64 {
	return this.count
}

func (this *Count) Write(p []byte) (int, error) {
	n, err := this.Writer.Write(p)
	this.count += int64(n)
	return n, err
}

/*



 */

type messageReader struct {
	buf  *bufio.Reader
	read func(buf *bufio.Reader) ([]byte, error)
}

func (this *messageReader) ReadMessage() ([]byte, error) {
	return this.read(this.buf)
}

//=============================

// MultiCloser 多个关闭合并
func MultiCloser(closer ...Closer) Closer {
	return &multiCloser{closer: closer}
}

type multiCloser struct {
	closer []Closer
}

func (this *multiCloser) Close() (err error) {
	for _, v := range this.closer {
		if er := v.Close(); er != nil {
			err = er
		}
	}
	return
}

//=============================

// PublisherToWriter Publisher to Writer
func PublisherToWriter(p Publisher, topic string) Writer {
	return &publishToWriter{topic: topic, Publisher: p}
}

type publishToWriter struct {
	topic string
	Publisher
}

func (this *publishToWriter) Write(p []byte) (int, error) {
	err := this.Publisher.Publish(this.topic, p)
	return len(p), err
}

//=============================

func NewReadWriter(r Reader, w Writer) ReadWriteCloser {
	return &readWriter{Reader: r, Writer: w}
}

type readWriter struct {
	Reader
	Writer
}

func (this *readWriter) Close() error { return nil }

//=============================

type Plan struct {
	Index   int    //操作次数
	Current int64  //当前数量
	Total   int64  //总数量
	Bytes   []byte //字节内容
}

type Connect interface {
	Closer
	Closed
	Runner

	// Connect 建立连接
	Connect(ctx context.Context) error

	// GetOnline 获取在线状态
	GetOnline() bool

	// SetOnline 设置在线状态,也可直接在Connect中实现
	// 方便统一管理,设置连接中,连接结果等
	SetOnline(online bool, reason string)
}
