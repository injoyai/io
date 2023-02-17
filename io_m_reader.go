package io

import (
	"bufio"
	"github.com/injoyai/io/buf"
	"io"
	"time"
)

func NewIReader(r Reader) *IReader {
	i := &IReader{
		buf:      bufio.NewReader(r),
		readFunc: buf.ReadWithAll,
		lastChan: make(chan Message),
	}
	if v, ok := r.(MessageReader); ok {
		i.readFunc = func(reader *bufio.Reader) ([]byte, error) {
			return v.ReadMessage()
		}
	}
	return i
}

type IReader struct {
	buf      *bufio.Reader                       //buffer
	readFunc func(*bufio.Reader) ([]byte, error) //读取函数
	lastChan chan Message                        //读取最新数据chan
	lastTime time.Time                           //最后读取数据时间
}

//================================Nature================================

// LastTime 最后数据时间
func (this *IReader) LastTime() time.Time {
	return this.lastTime
}

//================================Read================================

// Buffer 极大的增加读取速度
func (this *IReader) Buffer() *bufio.Reader {
	return this.buf
}

// Read io.reader
func (this *IReader) Read(p []byte) (int, error) {
	return this.Buffer().Read(p)
}

// ReadByte 读取一字节
func (this *IReader) ReadByte() (byte, error) {
	return this.Buffer().ReadByte()
}

// ReadAll 读取全部数据
func (this *IReader) ReadAll() ([]byte, error) {
	return buf.ReadWithAll(this.Buffer())
}

// ReadMessage 实现MessageReader接口
func (this *IReader) ReadMessage() ([]byte, error) {
	if this.readFunc == nil {
		return nil, ErrInvalidReadFunc
	}
	return this.readFunc(this.Buffer())
}

// ReadLast 读取最新的数据
func (this *IReader) ReadLast(timeout time.Duration) (response []byte, err error) {
	if timeout <= 0 {
		select {
		//case <-this.Ctx().Done():
		//	err = this.Err()
		case response = <-this.lastChan:
		}
	} else {
		t := time.NewTimer(timeout)
		select {
		//case <-this.ctx.Done():
		//	err = this.closeErr
		case response = <-this.lastChan:
		case <-t.C:
			err = ErrWithTimeout
		}
	}
	return
}

// WriteTo 写入io.Writer
func (this *IReader) WriteTo(writer Writer) (int64, error) {
	return Copy(writer, this)
}

// CopyTo 写入io.Writer
func (this *IReader) CopyTo(writer Writer) (int64, error) {
	return Copy(writer, this)
}

//================================ReadFunc================================

// SetReadFunc 设置读取函数
func (this *IReader) SetReadFunc(fn buf.ReadFunc) {
	this.readFunc = func(reader *bufio.Reader) ([]byte, error) {
		bs, err := fn(reader)
		if err != nil {
			return nil, err
		}
		if len(bs) > 0 {
			//设置最后读取有效数据时间
			this.lastTime = time.Now()
			//尝试加入通道
			select {
			case this.lastChan <- bs:
			default:
			}
		}
		return bs, nil
	}
}

// SetReadWithNil 设置读取函数为nil
func (this *IReader) SetReadWithNil() {
	this.SetReadFunc(nil)
}

// SetReadWithAll 一次性全部读取
func (this *IReader) SetReadWithAll() {
	this.SetReadFunc(buf.ReadWithAll)
}

// SetReadWithKB 读取固定字节长度
func (this *IReader) SetReadWithKB(n uint) {
	this.SetReadFunc(func(buf *bufio.Reader) ([]byte, error) {
		bytes := make([]byte, n<<10)
		length, err := buf.Read(bytes)
		return bytes[:length], err
	})
}

// SetReadWithStartEnd 设置根据包头包尾读取数据
func (this *IReader) SetReadWithStartEnd(packageStart, packageEnd []byte) {
	this.SetReadFunc(buf.NewReadWithStartEnd(packageStart, packageEnd))
}

// SetReadWithWriter same io.Copy 注意不能设置读取超时
func (this *IReader) SetReadWithWriter(writer io.Writer) {
	this.SetReadFunc(buf.NewReadWithWriter(writer))
}

// SetReadWithLenFrame 根据动态长度读取数据
func (this *IReader) SetReadWithLenFrame(f *buf.LenFrame) {
	this.SetReadFunc(buf.NewReadWithLen(f))
}

// SetReadWithTimeout 根据超时时间读取数据(需要及时读取,避免阻塞产生粘包)
func (this *IReader) SetReadWithTimeout(timeout time.Duration) {
	this.SetReadFunc(buf.NewReadWithTimeout(timeout))
}

// SetReadWithFrame 适配预大部分读取
func (this *IReader) SetReadWithFrame(f *buf.Frame) {
	this.SetReadFunc(buf.NewReadWithFrame(f))
}
