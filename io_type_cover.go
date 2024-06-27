package io

import (
	"encoding/base64"
	"encoding/hex"
	"github.com/injoyai/conv"
	"io"
	"sync/atomic"
)

//=============================覆盖读写=============================

// CoverWriter 覆盖写,写入经过handler处理后的数据
type CoverWriter struct {
	io.Writer
	Handler func(p []byte) ([]byte, error)
}

func (this *CoverWriter) Write(bs []byte) (n int, err error) {
	if this.Handler != nil {
		bs, err = this.Handler(bs)
		if err != nil {
			return 0, err
		}
	}
	return this.Writer.Write(bs)
}

// WriteString 写入字符串,实现io.StringWriter
func (this *CoverWriter) WriteString(s string) (int, error) {
	return this.Write([]byte(s))
}

// WriteHEX 写入16进制数据
func (this *CoverWriter) WriteHEX(s string) (int, error) {
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return 0, err
	}
	return this.Write(bytes)
}

// WriteBase64 写入base64数据
func (this *CoverWriter) WriteBase64(s string) (int, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return 0, err
	}
	return this.Write(bytes)
}

// WriteAny 写入任意数据,根据conv转成字节
func (this *CoverWriter) WriteAny(any interface{}) (int, error) {
	return this.Write(conv.Bytes(any))
}

// WriteSplit 写入字节,分片写入,例如udp需要写入字节小于(1500-20-8=1472)
func (this *CoverWriter) WriteSplit(p []byte, length int) (int, error) {
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

// WriteReader io.Reader
func (this *CoverWriter) WriteReader(reader Reader) (int64, error) {
	return Copy(this, reader)
}

// WriteChan 监听通道并写入
func (this *CoverWriter) WriteChan(c chan interface{}) (int64, error) {
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

//===================================================

// CoverReader 覆盖读,返回经过handler处理后的数据
type CoverReader struct {
	io.Reader
	Handler func(p []byte) ([]byte, error)
}

func (this *CoverReader) Read(bs []byte) (n int, err error) {
	n, err = this.Reader.Read(bs)
	if err != nil {
		return 0, err
	}
	if this.Handler != nil {
		bs, err = this.Handler(bs)
		if err != nil {
			return 0, err
		}
	}
	return
}

//===================================================

// CoverCloser 覆盖关闭，返回经过handler处理后的数据
type CoverCloser struct {
	io.Closer
	Handler func(err error)
	closed  uint32 //0是未关闭,1是关闭中,2是已关闭
}

func (this *CoverCloser) Close() error {
	return this.CloseWithErr(ErrHandClose)
}

func (this *CoverCloser) CloseWithErr(err error) error {
	if err == nil {
		return nil
	}
	if this.Closed() {
		return nil
	}
	if err := this.Closer.Close(); err != nil {
		return err
	}
	defer atomic.StoreUint32(&this.closed, 1)
	if this.Handler != nil {
		this.Handler(err)
	}
	return nil
}

// Closed 是否已关闭
func (this *CoverCloser) Closed() bool {
	return atomic.LoadUint32(&this.closed) > 0
}
