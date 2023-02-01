package proxy

import (
	"context"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"log"
	"time"
)

// NewClient 管道的客户端,监听客户端的消息,进行后续操作(建立连接,写数据,关闭连接)
func NewClient(addr, sn string) *Client {

	cli := &Client{sn: sn, addr: addr}

	//代理客户端,处理收到的信息
	cli.Proxy = func() *Proxy {
		proxy := NewProxy().SetKey(sn)
		proxy.SetConnectFunc(ConnectFunc(cli.Client))
		return proxy
	}()

	//读取到数据进行处理的队列,避免数据掺杂(例如同时发送)
	cli.dealQueue = func() *chans.Entity {
		queue := chans.NewEntity(20)
		queue.SetHandler(func(no, num int, data interface{}) {
			msg := data.(*io.ClientMessage)
			m, err := DecodeMsg(msg.Bytes())
			if err != nil {
				//不能确定是那个代理连接
				log.Printf("[错误]解析失败:%s", msg.String())
				return
			}
			if err := cli.Proxy.Switch(m); err != nil {
				log.Printf("[错误][%s] %s", msg.GetKey(), err.Error())
				cli.Client.Write(NewCloseMsg(m.Key, err.Error()).Bytes())
			}
		})
		return queue
	}()

	//通道客户端,与服务端(管道另一端)保持长连接
	cli.Client = func() *io.Client {
		return io.Redial(dial.TCPFunc(addr), func(ctx context.Context, c *io.Client) {
			cli.Proxy.CloseConnAll()
			c.SetPrintFunc(PrintFunc)
			c.SetWriteFunc(WriteFunc)
			c.SetReadFunc(ReadFunc)
			c.SetDealFunc(func(msg *io.ClientMessage) { cli.dealQueue.Do(msg) })
			c.SetPrintWithASCII()
			c.SetKey(sn)
			c.Write(NewRegisterMsg(sn, "").Bytes())
			c.SetKeepAlive(time.Minute, NewRegisterMsg(sn, "").Bytes())
		})
	}()

	return cli
}

// Client 通道客户端
type Client struct {
	*io.Client               //通道客户端
	sn         string        //客户端标识
	addr       string        //通道服务端地址
	dealQueue  *chans.Entity //处理读取到数据的执行队列
	Proxy      *Proxy        //代理客户端连接池
}

// Write 给服务端发数据 todo 数据太长进行分包(伪多线程),避免长期占用
func (this *Client) Write(msg []byte) (int, error) {
	return this.Client.Write(NewInfoMsg(this.sn, msg).Bytes())
}

// SetRedirectFunc 自定义重定向函数
func (this *Client) SetRedirectFunc(i IRedirect) *Client {
	this.Proxy.SetRedirect(i)
	return this
}

// Redirect 重定向,*表示重定向全部,用默认重定向函数是生效
func (this *Client) Redirect(oldAddr, newAddr string) *Client {
	this.Proxy.Redirect(oldAddr, newAddr)
	return this
}
