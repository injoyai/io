package io

import (
	"bufio"
	"context"
	"fmt"
	"github.com/injoyai/io/buf"
	"io"
	"sync/atomic"
	"time"
)

func NewIReadCloser(readCloser ReadCloser) *IReadCloser {
	return NewIReadCloserWithContext(context.Background(), readCloser)
}

func NewIReadCloserWithContext(ctx context.Context, readCloser ReadCloser) *IReadCloser {
	if c, ok := readCloser.(*IReadCloser); ok && c != nil {
		return c
	}
	entity := &IReadCloser{
		IPrinter: NewIPrinter(""),
		ICloser:  NewICloserWithContext(ctx, readCloser),
		buf:      bufio.NewReader(readCloser),
		readFunc: buf.ReadWithAll,
		running:  0,
		lastChan: make(chan Message),
	}
	entity.SetCloseFunc(func(ctx context.Context, msg Message) {
		entity.cancel()
	})
	return entity
}

type IReadCloser struct {
	*IPrinter
	*ICloser
	buf      *bufio.Reader                       //buffer
	readFunc func(*bufio.Reader) ([]byte, error) //读取函数
	dealFunc DealFunc                            //处理数据函数
	running  uint32                              //是否在运行
	lastTime time.Time                           //最后读取数据时间
	lastChan chan Message                        //读取最新数据chan
}

//================================Nature================================

func (this *IReadCloser) SetKey(key string) *IReadCloser {
	this.IPrinter.SetKey(key)
	this.ICloser.SetKey(key)
	return this
}

func (this *IReadCloser) SetPrintFunc(fn PrintFunc) *IReadCloser {
	this.IPrinter.SetPrintFunc(fn)
	//错误信息按ASCII编码
	return this
}

func (this *IReadCloser) Debug(b ...bool) *IReadCloser {
	this.IPrinter.Debug(b...)
	this.ICloser.Debug(b...)
	return this
}

// LastTime 最后数据时间
func (this *IReadCloser) LastTime() time.Time {
	return this.lastTime
}

// Buffer 极大的增加读取速度
func (this *IReadCloser) Buffer() *bufio.Reader {
	return this.buf
}

//================================ReadWrite================================

// Read io.reader
func (this *IReadCloser) Read(p []byte) (int, error) {
	return this.Buffer().Read(p)
}

// ReadByte 读取一字节
func (this *IReadCloser) ReadByte() (byte, error) {
	return this.Buffer().ReadByte()
}

// ReadAll 读取全部数据
func (this *IReadCloser) ReadAll() ([]byte, error) {
	return buf.ReadWithAll(this.Buffer())
}

// ReadLast 读取最新的数据
func (this *IReadCloser) ReadLast(timeout time.Duration) (response []byte, err error) {
	if timeout <= 0 {
		select {
		case <-this.Ctx().Done():
			err = this.Err()
		case response = <-this.lastChan:
		}
	} else {
		t := time.NewTimer(timeout)
		select {
		case <-this.ctx.Done():
			err = this.closeErr
		case response = <-this.lastChan:
		case <-t.C:
			err = ErrWithTimeout
		}
	}
	return
}

// WriteTo 写入io.Writer
func (this *IReadCloser) WriteTo(writer Writer) (int64, error) {
	return Copy(writer, this)
}

// CopyTo 写入io.Writer
func (this *IReadCloser) CopyTo(writer Writer) (int64, error) {
	return Copy(writer, this)
}

//================================ReadFunc================================

// SetReadFunc 设置读取函数
func (this *IReadCloser) SetReadFunc(fn buf.ReadFunc) {
	this.readFunc = fn
}

// SetReadWithNil 设置读取函数为nil
func (this *IReadCloser) SetReadWithNil() {
	this.SetReadFunc(nil)
}

// SetReadWithAll 一次性全部读取
func (this *IReadCloser) SetReadWithAll() {
	this.SetReadFunc(buf.ReadWithAll)
}

// SetReadWithKB 读取固定字节长度
func (this *IReadCloser) SetReadWithKB(n uint) {
	this.SetReadFunc(func(buf *bufio.Reader) ([]byte, error) {
		bytes := make([]byte, n<<10)
		length, err := buf.Read(bytes)
		return bytes[:length], err
	})
}

// SetReadWithStartEnd 设置根据包头包尾读取数据
func (this *IReadCloser) SetReadWithStartEnd(packageStart, packageEnd []byte) {
	this.SetReadFunc(buf.NewReadWithStartEnd(packageStart, packageEnd))
}

// SetReadWithWriter same io.Copy 注意不能设置读取超时
func (this *IReadCloser) SetReadWithWriter(writer io.Writer) {
	this.SetReadFunc(buf.NewReadWithWriter(writer))
}

// SetReadWithLenFrame 根据动态长度读取数据
func (this *IReadCloser) SetReadWithLenFrame(f *buf.LenFrame) {
	this.SetReadFunc(buf.NewReadWithLen(f))
}

// SetReadWithTimeout 根据超时时间读取数据(需要及时读取,避免阻塞产生粘包)
func (this *IReadCloser) SetReadWithTimeout(timeout time.Duration) {
	this.SetReadFunc(buf.NewReadWithTimeout(timeout))
}

// SetReadWithFrame 适配预大部分读取
func (this *IReadCloser) SetReadWithFrame(f *buf.Frame) {
	this.SetReadFunc(buf.NewReadWithFrame(f))
}

//================================DealFunc================================

// SetDealFunc 设置数据处理函数
func (this *IReadCloser) SetDealFunc(fn func(msg Message)) {
	this.dealFunc = fn
}

// SetDealWithNil 不设置数据处理函数
func (this *IReadCloser) SetDealWithNil() {
	this.SetDealFunc(nil)
}

// SetDealWithWriter 设置数据处理到io.Writer
func (this *IReadCloser) SetDealWithWriter(writer Writer) {
	this.SetDealFunc(func(msg Message) {
		writer.Write(msg)
	})
}

// SetDealWithChan 设置数据处理到chan
func (this *IReadCloser) SetDealWithChan(c chan Message) {
	this.SetDealFunc(func(msg Message) {
		c <- msg
	})
}

//================================RunTime================================

// Running 是否在运行
func (this *IReadCloser) Running() bool {
	return this.running == 1
}

func (this *IReadCloser) Run() error {
	if atomic.SwapUint32(&this.running, 1) == 1 {
		return nil
	}
	for {
		select {
		case <-this.Done():
			return this.Err()
		default:
			_ = this.CloseWithErr(func() (err error) {
				defer func() {
					if e := recover(); e != nil {
						err = fmt.Errorf("%v", e)
					}
				}()
				if this.readFunc == nil {
					return ErrInvalidReadFunc
				}
				bytes, err := this.readFunc(this.Buffer())
				if err != nil || len(bytes) == 0 {
					return err
				}
				//设置最后读取有效数据时间
				this.lastTime = time.Now()
				//打印日志
				this.IPrinter.Print(bytes, TagRead, this.GetKey())
				select {
				case this.lastChan <- bytes:
					//尝试加入队列
				default:
				}
				//处理数据
				if this.dealFunc != nil {
					this.dealFunc(bytes)
				}
				return nil
			}())
		}
	}
}
