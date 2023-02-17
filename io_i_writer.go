package io

import (
	"encoding/base64"
	"encoding/hex"
	"github.com/injoyai/conv"
	"time"
)

// NewWriter 新建写
func NewWriter(writer Writer) *IWriter {
	if c, ok := writer.(*IWriter); ok && c != nil {
		return c
	}
	return &IWriter{
		IPrinter:  NewIPrinter(""),
		writer:    writer,
		writeFunc: nil,
		lastTime:  time.Time{},
	}
}

type IWriter struct {
	*IPrinter                       //打印
	writer    Writer                //io.Writer
	writeFunc func(p []byte) []byte //写入函数
	lastTime  time.Time             //最后写入时间
}

//================================Nature================================

// LastTime 最后数据时间
func (this *IWriter) LastTime() time.Time {
	return this.lastTime
}

// Write 写入字节,实现io.Writer
func (this *IWriter) Write(p []byte) (int, error) {
	if this.writeFunc != nil {
		p = this.writeFunc(p)
	}
	num, err := this.writer.Write(p)
	if err != nil {
		return 0, dealErr(err)
	}
	this.lastTime = time.Now()
	this.Print(NewMessage(p), TagWrite, this.GetKey())
	return num, nil
}

// WriteBytes 写入字节,实现bytesWriter
func (this *IWriter) WriteBytes(p []byte) (int, error) {
	return this.Write(p)
}

// WriteString 写入字符串,实现io.StringWriter
func (this *IWriter) WriteString(s string) (int, error) {
	return this.Write([]byte(s))
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

// WriteWithTimeout 写入或者超时,todo 待实现
func (this *IWriter) WriteWithTimeout(p []byte, timeout time.Duration) (int, error) {
	return this.Write(p)
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

//================================WriteFunc================================

// SetWriteFunc 设置写入函数
func (this *IWriter) SetWriteFunc(fn func(p []byte) []byte) {
	this.writeFunc = fn
}

// SetWriteWithNil 取消写入函数
func (this *IWriter) SetWriteWithNil() {
	this.writeFunc = nil
}

// SetWriteWithStartEnd 设置写入函数,增加头尾
func (this *IWriter) SetWriteWithStartEnd(start, end []byte) {
	this.writeFunc = func(p []byte) []byte {
		return append(start, append(p, end...)...)
	}
}
