package io

import (
	"bufio"
	"bytes"
	"context"
	"github.com/injoyai/io/buf"
	"sync/atomic"
	"time"
)

//================================Nature================================

// Buffer 极大的增加读取速度
func (this *Client) Buffer() *bufio.Reader {
	return this.buf
}

// SetBufferSize 设置缓存大小
func (this *Client) SetBufferSize(size int) error {
	old := make([]byte, this.buf.Size())
	n, err := this.buf.Read(old)
	if err != nil {
		return err
	}
	reader := MultiReader(bytes.NewReader(old[:n]), this.i)
	this.buf = bufio.NewReaderSize(reader, size)
	return nil
}

// Read io.reader
func (this *Client) Read(p []byte) (int, error) {
	return this.Buffer().Read(p)
}

// ReadByte 读取一字节
func (this *Client) ReadByte() (byte, error) {
	return this.Buffer().ReadByte()
}

// Read1KB 读取1kb数据
func (this *Client) Read1KB() ([]byte, error) {
	return buf.Read1KB(this.Buffer())
}

// ReadMessage 实现MessageReader接口
func (this *Client) ReadMessage() ([]byte, error) {
	ack, err := this.ReadAck()
	if err != nil {
		return nil, err
	}
	return ack.Payload(), nil
}

func (this *Client) ReadAck() (Acker, error) {
	if this.readFunc == nil {
		return nil, ErrInvalidReadFunc
	}
	return this.readFunc(this.Buffer())
}

// ReadLatest 读取最新的数据
func (this *Client) ReadLatest(timeout time.Duration) (response []byte, err error) {
	if timeout <= 0 {
		response = <-this.latestChan
	} else {
		select {
		case response = <-this.latestChan:
		case <-time.After(timeout):
			err = ErrWithTimeout
		}
	}
	return
}

// WriteTo 写入io.Writer
func (this *Client) WriteTo(writer Writer) (int64, error) {
	return Copy(writer, this)
}

// SetReadTimeout 设置读取间隔超时时间,需要在Run之前设置
func (this *Client) SetReadTimeout(timeout time.Duration) *Client {
	this.timeout = timeout
	return this
}

//================================RunTime================================

// Running 是否在运行
func (this *Client) Running() bool {
	return atomic.LoadUint32(&this.running) == 1
}

// Run 开始运行数据读取
// 2个操作,持续读取数据并处理,监测超时(如果设置了超时,不推荐服务端使用)
func (this *Client) Run() error {

	//原子操作,防止重复执行
	if atomic.SwapUint32(&this.running, 1) == 1 {
		return nil
	}

	//todo is a good idea ?
	if this.timeout > 0 {
		go func(ctx context.Context) {
			timer := time.NewTimer(this.timeout)
			defer timer.Stop()
			for {
				timer.Reset(this.timeout)
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					_ = this.CloseWithErr(ErrWithReadTimeout)
					return
				case <-this.timeoutReset:
				}
			}
		}(this.Ctx())
	}

	//开始循环读取数据,处理数据
	return this.For(func(ctx context.Context) (err error) {

		//读取数据
		ack, err := this.ReadAck()
		if err != nil || len(ack.Payload()) == 0 {
			return err
		}

		//处理数据
		for _, dealFunc := range this.dealFunc {
			if dealFunc != nil && dealFunc(this, ack.Payload()) {
				ack.Ack()
			}
		}

		return nil
	})

}
