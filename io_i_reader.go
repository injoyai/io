package io

import (
	"bufio"
	"github.com/injoyai/io/buf"
	"io"
	"time"
)

// NewIReader 新建IReader,默认读取函数ReadAll
func NewIReader(r Reader) *IReader {
	i := &IReader{
		Key:      "",
		Logger:   defaultLogger(),
		lastChan: make(chan Message),
		lastTime: time.Now(),
	}
	if v, ok := r.(MessageReader); ok {
		i.mReader = v
	} else {
		i.buf = bufio.NewReaderSize(r, 1<<20)
	}
	i.SetReadFunc(buf.ReadWithAll)
	return i
}

type IReader struct {
	Key
	Logger     *logger
	mReader    MessageReader                           //接口MessageReader,兼容Reader
	buf        *bufio.Reader                           //buffer
	readFunc   func(buf *bufio.Reader) ([]byte, error) //读取函数
	lastChan   chan Message                            //读取最新数据chan
	lastTime   time.Time                               //最后读取数据时间
	bytesCount int64                                   //读取的字节数
}

//================================Nature================================

// LastTime 最后数据时间
func (this *IReader) LastTime() time.Time {
	return this.lastTime
}

// BytesCount 写入的字节数
func (this *IReader) BytesCount() int64 {
	return this.bytesCount
}

// Buffer 极大的增加读取速度
func (this *IReader) Buffer() *bufio.Reader {
	return this.buf
}

//================================Read================================

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
		response = <-this.lastChan
	} else {
		select {
		case response = <-this.lastChan:
		case <-time.After(timeout):
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
func (this *IReader) SetReadFunc(fn func(*bufio.Reader) ([]byte, error)) *IReader {
	this.readFunc = func(reader *bufio.Reader) (bs []byte, err error) {
		switch true {
		case this.mReader != nil:
			//特殊处理MessageReader
			bs, err = this.mReader.ReadMessage()
		case fn == nil:
			//默认读取全部
			bs, err = buf.ReadWithAll(reader)
		default:
			//按用户设置函数
			bs, err = fn(reader)
		}
		if err != nil {
			return nil, err
		}
		if len(bs) > 0 {
			//设置最后读取有效数据时间
			this.lastTime = time.Now()
			this.bytesCount += int64(len(bs))
			//尝试加入通道
			select {
			case this.lastChan <- bs:
			default:
			}
			//打印日志
			this.Logger.Readln("["+this.GetKey()+"] ", bs)
		}
		return bs, nil
	}
	return this
}

// SetReadWithPkg 使用默认读包方式
func (this *IReader) SetReadWithPkg() *IReader {
	return this.SetReadFunc(ReadWithPkg)
}

// SetReadWithAll 一次性全部读取
func (this *IReader) SetReadWithAll() *IReader {
	return this.SetReadFunc(buf.ReadWithAll)
}

// SetReadWithKB 读取固定字节长度
func (this *IReader) SetReadWithKB(n uint) *IReader {
	return this.SetReadFunc(func(buf *bufio.Reader) ([]byte, error) {
		bytes := make([]byte, n<<10)
		length, err := buf.Read(bytes)
		return bytes[:length], err
	})
}

// SetReadWithStartEnd 设置根据包头包尾读取数据
func (this *IReader) SetReadWithStartEnd(packageStart, packageEnd []byte) *IReader {
	return this.SetReadFunc(buf.NewReadWithStartEnd(packageStart, packageEnd))
}

// SetReadWithWriter same io.Copy 注意不能设置读取超时
func (this *IReader) SetReadWithWriter(writer io.Writer) *IReader {
	return this.SetReadFunc(buf.NewReadWithWriter(writer))
}

// Bridge 桥接模式,等同SetReadWithWriter
func (this *IReader) Bridge(w ...io.Writer) *IReader {
	return this.SetReadFunc(buf.NewReadWithWriter(MultiWriter(w...)))
}

// SetReadWithLenFrame 根据动态长度读取数据
func (this *IReader) SetReadWithLenFrame(f *buf.LenFrame) *IReader {
	return this.SetReadFunc(buf.NewReadWithLen(f))
}

// SetReadWithTimeout 根据超时时间读取数据(需要及时读取,避免阻塞产生粘包)
func (this *IReader) SetReadWithTimeout(timeout time.Duration) *IReader {
	return this.SetReadFunc(buf.NewReadWithTimeout(timeout))
}

// SetReadWithFrame 适配预大部分读取
func (this *IReader) SetReadWithFrame(f *buf.Frame) *IReader {
	return this.SetReadFunc(buf.NewReadWithFrame(f))
}
