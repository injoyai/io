package io

import (
	"github.com/injoyai/io/buf"
)

// multiCloser
// 合并多个Closer , 变成1个Closer
type multiCloser struct {
	closer []Closer
}

func (this *multiCloser) Close() (err error) {
	for _, v := range this.closer {
		if er := v.Close(); er != nil {
			err = er
		}
	}
	return
}

// MultiCloser 多个关闭合并
func MultiCloser(closer ...Closer) Closer {
	return &multiCloser{closer: closer}
}

type publishToWriter struct {
	topic string
	Publisher
}

func (this *publishToWriter) Write(p []byte) (int, error) {
	err := this.Publisher.Publish(this.topic, p)
	return len(p), err
}

// PublisherToWriter Publisher to Writer
func PublisherToWriter(p Publisher, topic string) Writer {
	return &publishToWriter{topic: topic, Publisher: p}
}

// SwapClient 数据交换交换
func SwapClient(c1, c2 *Client) {
	c1.SetDealWithWriter(c2)
	c2.SetDealWithWriter(c1)
	go c1.Run()
	go c2.Run()
}

// SwapWithReadFunc 根据读取规则来交换数据(例如数据进行了加密),需要返回解密的字节
func SwapWithReadFunc(i1, i2 ReadWriteCloser, readFunc buf.ReadFunc) {
	c1 := NewClient(i1)
	c1.SetReadFunc(readFunc)
	c2 := NewClient(i2)
	c2.SetReadFunc(readFunc)
	SwapClient(c1, c2)
}

// Swap same two Copy IO数据交换
func Swap(i1, i2 ReadWriteCloser) {
	c1 := NewClient(i1)
	c1.SetReadWithAll()
	c2 := NewClient(i2)
	c2.SetReadWithAll()
	SwapClient(c1, c2)
}
