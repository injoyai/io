package io

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"github.com/injoyai/conv"
	"github.com/injoyai/io/buf"
	"time"
)

const (
	writeQueueKey = "_write_queue"
)

// NewIWriter 新建写
func NewIWriter(writer Writer) *IWriter {
	if c, ok := writer.(*IWriter); ok && c != nil {
		return c
	}
	return &IWriter{
		Key:      &Key{},
		Logger:   NewLog(),
		writer:   writer,
		lastTime: time.Time{},
	}
}

// IWriter 写
type IWriter struct {
	*Key
	Logger    Logger
	writer    Writer    //io.Writer
	writeFunc WriteFunc //写入函数
	lastTime  time.Time //最后写入时间
}

//================================Nature================================

// LastTime 最后数据时间
func (this *IWriter) LastTime() time.Time {
	return this.lastTime
}

// Write 写入字节,实现io.Writer
func (this *IWriter) Write(p []byte) (n int, err error) {
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
	return this.WriteASCII(s)
}

// WriteASCII 写入ascii码数据
func (this *IWriter) WriteASCII(s string) (int, error) {
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

// NewWriteQueue 新建写入队列
func (this *IWriter) NewWriteQueue(ctx context.Context, length ...int) chan []byte {
	queue := make(chan []byte, conv.GetDefaultInt(100, length...))
	go func(ctx context.Context) {
		//defer close(queue) 自动回收
		for {
			select {
			case <-ctx.Done():
				return
			case p := <-queue:
				this.Write(p)
			}
		}
	}(ctx)
	return queue
}
