package io

import (
	"bytes"
	"errors"
	"io"
)

func NewVirtual(key string, w io.Writer) *Virtual {
	return &Virtual{
		Key:    key,
		Writer: w,
		Reader: bytes.NewBuffer(nil),
	}
}

// Virtual 虚拟设备
// 在
type Virtual struct {
	Key     string                       //标识
	Writer  io.Writer                    //tun连接
	Reader  *bytes.Buffer                //缓存
	closed  bool                         //判断是否被关闭
	OnWrite func([]byte) ([]byte, error) //写入事件
	OnClose func(v *Virtual) error       //关闭事件
}

func (this *Virtual) Read(p []byte) (n int, err error) {
	if this.Reader.Len() == 0 && this.closed {
		return 0, io.EOF
	}
	return this.Reader.Read(p)
}

func (this *Virtual) Write(p []byte) (n int, err error) {
	if this.closed {
		return 0, errors.New("closed")
	}
	return this.Writer.Write(p)
}

func (this *Virtual) Close() error {
	if this.closed {
		return nil
	}
	this.closed = true
	return nil
}
