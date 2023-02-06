package pipe

import (
	"context"
	"github.com/injoyai/io"
	"log"
)

// NewClient 新建管道客户端
func NewClient(dial io.DialFunc) (*Client, error) {
	client, err := io.NewDial(dial)
	if err != nil {
		return nil, err
	}
	c := &Client{
		Client:   client,
		writeLen: 2 << 10,
		dealFunc: func(msg *Message) error {
			log.Printf("[Client]%s\n", string(msg.Data))
			return nil
		},
	}
	c.Redial()
	return c, nil
}

/*
Client
抽象管道概念
例如 用485线通讯,正常的TCP连接 都属于管道
需要 客户端对客户端 客户端对服务端 2种方式
需要 一个管道通讯多个io数据,并且不能长期占用(交替使用,伪多线程)

提供io.Reader io.Writer接口
写入数据会封装一层(封装连接信息,动作,数据)

*/
type Client struct {

	//io客户端
	*io.Client

	//分包长度,每次写入固定字节长度,默认1k
	//0表示一次性全部写入
	//数据太多可能会影响到其他数据的时效
	writeLen uint

	////数据缓存,等到数据被读取
	////单通道支持不同连接的数据
	//mapIO map[string]io.WriteCloser
	//
	//mu sync.RWMutex

	//处理读取到的数据
	dealFunc func(msg *Message) error
}

// SetDealFunc 设置处理函数
func (this *Client) SetDealFunc(fn func(msg *Message) error) *Client {
	this.dealFunc = fn
	return this
}

//func (this *Client) SetIO(key string, writeCloser io.WriteCloser) *Client {
//	this.mu.RLock()
//	oldIO := this.mapIO[key]
//	this.mu.RUnlock()
//	if oldIO != nil && oldIO != writeCloser {
//		oldIO.Close()
//	}
//	this.mu.Lock()
//	defer this.mu.Unlock()
//	this.mapIO[key] = writeCloser
//	return this
//}

// Redial 重连初始化
func (this *Client) Redial(fn ...func(ctx context.Context, c *io.Client)) {
	this.Client.Redial(func(ctx context.Context, c *io.Client) {
		this.Client.SetWriteFunc(encodePackage)
		this.Client.SetReadFunc(defaultReadFunc)
		this.Client.SetDealFunc(newDealFunc(this.dealFunc))
		for _, v := range fn {
			v(ctx, c)
		}

		////写入数据到缓存
		//this.mu.RLock()
		//writeCloser := this.mapIO[result.Key]
		//this.mu.RUnlock()
		//if writeCloser != nil {
		//
		//	//如果远程已关闭,则断开当前连接,并删除
		//	if string(result.Data) == io.EOF.Error() {
		//		writeCloser.Close()
		//		this.mu.Lock()
		//		delete(this.mapIO, result.Key)
		//		this.mu.Unlock()
		//		return
		//	}
		//
		//	//存在则写入数据
		//	writeCloser.Write(result.Data)
		//	return
		//
		//}
		//
		////不存在则说明连接已经关闭,或者远程已关闭
		////下发关闭数据到通道另一头
		//m := newCloseMessage(result.Key, "连接已关闭")
		//this.Client.Write(encodePackage(m.Bytes()))

	})
}

// Write 实现io.Writer
func (this *Client) Write(key, addr string, p []byte) (int, error) {
	return writeFunc(this.writeLen, this.Client.Write, key, addr, p)
}

func (this *Client) WriteMessage(msg *Message) error {
	_, err := this.Client.Write(msg.Bytes())
	return err
}
