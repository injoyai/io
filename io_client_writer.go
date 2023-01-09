package io

import (
	"time"
)

type ClientWriter struct {
	*ClientPrint
	writer    Writer                //io.Writer
	writeFunc func(p []byte) []byte //写入函数
	lastTime  time.Time             //最后写入时间
}

// Write 写入字节,实现io.Writer
func (this *ClientWriter) Write(p []byte) (int, error) {
	if this.writeFunc != nil {
		p = this.writeFunc(p)
	}
	if this.printFunc != nil {
		this.printFunc("发送", NewMessage(p))
	}
	return this.writer.Write(p)
}

// WriteBytes 写入字节,实现bytesWriter
func (this *ClientWriter) WriteBytes(p []byte) (int, error) {
	return this.Write(p)
}

// WriteString 写入字符串,实现io.StringWriter
func (this *ClientWriter) WriteString(s string) (int, error) {
	return this.Write([]byte(s))
}

func (this *ClientWriter) WriteWithTimeout(p []byte, timeout time.Duration) (int, error) {
	return this.Write(p)
}

// SetWriteFunc 设置写入函数
func (this *ClientWriter) SetWriteFunc(fn func(p []byte) []byte) {
	this.writeFunc = fn
}

// SetWriteWithNil 取消写入函数
func (this *ClientWriter) SetWriteWithNil() {
	this.writeFunc = nil
}

// SetWriteWithStartEnd 设置写入函数,增加头尾
func (this *ClientWriter) SetWriteWithStartEnd(start, end []byte) {
	this.writeFunc = func(p []byte) []byte {
		return append(start, append(p, end...)...)
	}
}
