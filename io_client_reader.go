package io

import (
	"bufio"
	"context"
	"github.com/injoyai/io/buf"
	"io"
	"sync/atomic"
	"time"
)

func NewClientReader(reader Reader) *ClientReader {
	return NewClientReaderWithContext(context.Background(), reader)
}

func NewClientReaderWithContext(ctx context.Context, reader Reader) *ClientReader {
	ctx, cancel := context.WithCancel(ctx)
	return &ClientReader{
		buf:      bufio.NewReader(reader),
		readChan: make(chan []byte),
		readFunc: buf.ReadWithAll,
		ctx:      ctx,
		cancel:   cancel,
		running:  &atomic.Value{},
	}
}

type ClientReader struct {
	*ClientPrint
	buf      *bufio.Reader                       //buff
	readChan chan []byte                         //读取数据chan
	readFunc func(*bufio.Reader) ([]byte, error) //读取函数
	dealFunc func(*Message)                      //处理数据函数
	ctx      context.Context                     //上下文
	cancel   context.CancelFunc                  //上下文关闭
	running  *atomic.Value                       //是否在运行
	lastTime time.Time                           //最后读取数据时间
}

// Buffer 极大的增加读取速度
func (this *ClientReader) Buffer() *bufio.Reader {
	return this.buf
}

// Read io.reader
func (this *ClientReader) Read(p []byte) (int, error) {
	return this.Buffer().Read(p)
}

// ReadByte 读取一字节
func (this *ClientReader) ReadByte() (byte, error) {
	return this.Buffer().ReadByte()
}

// ReadAll 读取全部数据
func (this *ClientReader) ReadAll() ([]byte, error) {
	return buf.ReadWithAll(this.Buffer())
}

//================================ReadFunc================================

// SetReadFunc 设置读取函数
func (this *ClientReader) SetReadFunc(fn func(c *bufio.Reader) ([]byte, error)) {
	this.readFunc = fn
}

// SetReadWithNil 设置读取函数为nil
func (this *ClientReader) SetReadWithNil() {
	this.SetReadFunc(nil)
}

// SetReadWithAll 一次性全部读取
func (this *ClientReader) SetReadWithAll() {
	this.SetReadFunc(buf.ReadWithAll)
}

// SetReadWithStartEnd 设置根据包头包尾读取数据
func (this *ClientReader) SetReadWithStartEnd(packageStart, packageEnd []byte) {
	this.SetReadFunc(buf.NewReadWithStartEnd(packageStart, packageEnd))
}

// SetReadWithWriter same io.Copy 注意不能设置读取超时
func (this *ClientReader) SetReadWithWriter(writer io.Writer) {
	this.SetReadFunc(buf.NewReadWithWriter(writer))
}

// SetReadWithLenFrame 根据动态长度读取数据
func (this *ClientReader) SetReadWithLenFrame(f *buf.LenFrame) {
	this.SetReadFunc(buf.NewReadWithLen(f))
}

// SetReadWithTimeout 根据超时时间读取数据(需要及时读取,避免阻塞产生粘包)
func (this *ClientReader) SetReadWithTimeout(timeout time.Duration) {
	this.SetReadFunc(buf.NewReadWithTimeout(timeout))
}

// SetReadWithFrame 适配预大部分读取
func (this *ClientReader) SetReadWithFrame(f *buf.Frame) {
	this.SetReadFunc(buf.NewReadWithFrame(f))
}

//================================DealFunc================================

// SetDealFunc 设置数据处理函数
func (this *ClientReader) SetDealFunc(fn func(msg *Message)) {
	this.dealFunc = fn
}

// SetDealWithNil 不设置数据处理函数
func (this *ClientReader) SetDealWithNil() {
	this.SetDealFunc(nil)
}

// SetDealWithWriter 设置数据处理到io.Writer
func (this *ClientReader) SetDealWithWriter(writer Writer) {
	this.SetDealFunc(func(msg *Message) {
		writer.Write(msg.Bytes())
	})
}

// Running 是否在运行,原子操作
func (this *ClientReader) Running() bool {
	v := this.running.Load()
	return v != nil && v.(bool)
}

// Close 关闭
func (this *ClientReader) Close() error {
	if this.cancel != nil {
		this.cancel()
	}
	return nil
}

func (this *ClientReader) Run() error {
	if this.Running() {
		select {
		case <-this.ctx.Done():
			return this.ctx.Err()
		}
	}
	for {
		select {
		case <-this.ctx.Done():
			return this.ctx.Err()
		default:
			if this.readFunc == nil {
				return ErrInvalidReadFunc
			}
			bytes, err := this.readFunc(this.Buffer())
			if err != nil {
				return err
			}
			//设置最后读取有效数据时间
			this.lastTime = time.Now()
			select {
			case this.readChan <- bytes:
				//尝试加入队列
			default:
			}
			if this.dealFunc != nil {
				this.dealFunc(NewMessage(bytes))
			}
		}
	}
}

//type ClientReader interface {
//	Reader
//	Buffer() *bufio.Reader
//	ReadByte() (byte, error)
//	ReadAll() (bytes []byte, err error)
//	SetReadWithStartEnd(packageStart, packageEnd []byte)
//	SetReadWithWriter(writer io.Writer)
//	SetReadWithFrame(f *buf.Frame)
//	SetReadWithNil()
//	SetReadFunc(fn func(c *bufio.Reader) ([]byte, error))
//	SetReadWithAll()
//}
//
//type ClientWriter interface {
//	Writer
//	StringWriter
//	BytesWriter
//	TimeoutWriter
//	SetWriteFunc(fn func([]byte) []byte)
//	SetWriteWithNil()
//}
//
//type ClientCloser interface {
//	Closer
//	Closed
//	SetCloseFunc(fn func(msg *bytes.Buffer))
//	SetCloseWithNil()
//}
//
//type ClientPrint interface {
//	SetPrintFunc(fn func(s string, msg *bytes.Buffer))
//	SetPrintWithHEX()
//	SetPrintWithASCII()
//}
//
//type ClientDeal interface {
//	SetDealFunc(fn func(msg *bytes.Buffer))
//	SetDealWithNil()
//	SetDealWithPrintASCII()
//	SetDealWithPrintHEX()
//}
//
//type Client interface {
//	ClientReader
//	ClientWriter
//	ClientCloser
//	ClientPrint
//
//	Debugger
//
//	Tag() *maps.Safe
//	GetTag(key interface{}) interface{}
//	SetTag(key, value interface{})
//	SetKey(key string)
//	GetKey() string
//	Err() error
//	// Run 运行,会阻塞
//	Run() error
//
//	Ctx() context.Context
//	Running() bool
//
//	// SetTimeout 设置超时时间
//	SetTimeout(timeout time.Duration)
//
//	ReadChan() <-chan []byte
//
//	ReadLast(timeout time.Duration) (response []byte, err error)
//
//	WriteRead(request []byte) (response []byte, err error)
//
//	WriteReadWithTimeout(request []byte, timeout time.Duration) (response []byte, err error)
//
//	SetKeepAlive(t time.Duration, keeps ...[]byte)
//
//	SetRedialFunc(fn func() (ReadWriteCloser, error))
//
//	Redial(fn ...func(c *Client))
//	SetRunType(runType string)
//}
