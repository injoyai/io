package io

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/injoyai/conv"
	"github.com/injoyai/io/buf"
	"sync"
	"time"
)

// NewIWriter 新建写
func NewIWriter(writer Writer) *IWriter {
	if c, ok := writer.(*IWriter); ok && c != nil {
		return c
	}
	return &IWriter{
		Key:      "",
		Logger:   defaultLogger(),
		writer:   writer,
		lastTime: time.Time{},
	}
}

// IWriter 写
type IWriter struct {
	Key
	Logger     *logger
	writer     Writer                         //io.Writer
	writeFunc  func(p []byte) ([]byte, error) //写入函数,处理写入内容
	err        error                          //写入错误信息
	queue      chan []byte                    //写入队列
	queueOnce  sync.Once                      //队列初始化
	lastTime   time.Time                      //最后写入时间
	bytesCount int64                          //写入的字节数
}

//================================Nature================================

// SetLogger 设置日志
func (this *IWriter) SetLogger(logger Logger) *IWriter {
	this.Logger = NewLogger(logger)
	return this
}

// LastTime 最后数据时间
func (this *IWriter) LastTime() time.Time {
	return this.lastTime
}

// BytesCount 写入的字节数
func (this *IWriter) BytesCount() int64 {
	return this.bytesCount
}

//================================Write================================

// Write 写入字节,实现io.Writer
func (this *IWriter) Write(p []byte) (n int, err error) {
	//记录错误,用于队列的判断
	defer func() {
		if err != nil {
			this.err = err
			if this.queue != nil {
				for {
					select {
					case <-this.queue:
					default:
						return
					}
				}
			}
		}
	}()
	if this.writeFunc != nil {
		p, err = this.writeFunc(p)
		if err != nil {
			return 0, dealErr(err)
		}
	}
	//打印实际发送的数据,方便调试
	this.Logger.Writeln("["+this.GetKey()+"] ", p)
	//写入数据
	n, err = this.writer.Write(p)
	if err != nil {
		return 0, dealErr(err)
	}
	this.lastTime = time.Now()
	this.bytesCount += int64(n)
	return
}

// WriteBytes 写入字节,实现bytesWriter
func (this *IWriter) WriteBytes(p []byte) (int, error) {
	return this.Write(p)
}

// WriteSplit 写入字节,分片写入,例如udp需要写入字节小于(1500-20-8=1472)
func (this *IWriter) WriteSplit(p []byte, length int) (int, error) {
	if length <= 0 {
		return this.Write(p)
	}
	for len(p) > 0 {
		var data []byte
		if len(p) >= length {
			data, p = p[:length], p[length:]
		} else {
			data, p = p, p[:0]
		}
		_, err := this.Write(data)
		if err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

// WriteString 写入字符串,实现io.StringWriter
func (this *IWriter) WriteString(s string) (int, error) {
	return this.Write([]byte(s))
}

// WriteHEX 写入16进制数据
func (this *IWriter) WriteHEX(s string) (int, error) {
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return 0, err
	}
	return this.Write(bytes)
}

// WriteBase64 写入base64数据
func (this *IWriter) WriteBase64(s string) (int, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return 0, err
	}
	return this.Write(bytes)
}

// WriteAny 写入任意数据,根据conv转成字节
func (this *IWriter) WriteAny(any interface{}) (int, error) {
	return this.Write(conv.Bytes(any))
}

// WriteReader io.Reader
func (this *IWriter) WriteReader(reader Reader) (int64, error) {
	return Copy(this, reader)
}

// Copy io.Reader
func (this *IWriter) Copy(reader Reader) (int64, error) {
	return Copy(this, reader)
}

// WriteChan 监听通道并写入
func (this *IWriter) WriteChan(c chan interface{}) (int64, error) {
	var total int64
	for data := range c {
		n, err := this.Write(conv.Bytes(data))
		if err != nil {
			return 0, err
		}
		total += int64(n)
	}
	return total, nil
}

//================================Queue================================

func (this *IWriter) WriteQueue(p []byte) (int, error) {
	if this.err != nil {
		return 0, this.err
	}
	this.initQueue()
	this.queue <- p
	return len(p), nil
}

func (this *IWriter) WriteQueueTry(p []byte) (int, error) {
	if this.err != nil {
		return 0, this.err
	}
	this.initQueue()
	select {
	case this.queue <- p:
		return len(p), nil
	default:
		return 0, errors.New("队列已满")
	}
}

func (this *IWriter) WriteQueueTimeout(p []byte, timeout time.Duration) (int, error) {
	if this.err != nil {
		return 0, this.err
	}
	this.initQueue()
	select {
	case this.queue <- p:
		return len(p), nil
	case <-time.After(timeout):
		return 0, ErrWithWriteTimeout
	}
}

func (this *IWriter) initQueue() {
	this.queueOnce.Do(func() {
		this.queue = make(chan []byte, DefaultChannelSize)
		go func() {
			//defer close(queue) 自动回收
			for p := range this.queue {
				if _, err := this.Write(p); err != nil {
					return
				}
			}
		}()
	})
}

//================================WriteFunc================================

// SetWriteFunc 设置写入函数,封装数据包,same SetWriteBeforeFunc
func (this *IWriter) SetWriteFunc(fn func(p []byte) ([]byte, error)) *IWriter {
	this.writeFunc = fn
	return this
}

// SetWriteWithPkg 默认写入函数
func (this *IWriter) SetWriteWithPkg() *IWriter {
	return this.SetWriteFunc(WriteWithPkg)
}

// SetWriteWithNil 取消写入函数
func (this *IWriter) SetWriteWithNil() *IWriter {
	return this.SetWriteFunc(nil)
}

// SetWriteWithStartEnd 设置写入函数,增加头尾
func (this *IWriter) SetWriteWithStartEnd(start, end []byte) *IWriter {
	return this.SetWriteFunc(buf.NewWriteWithStartEnd(start, end))
}
