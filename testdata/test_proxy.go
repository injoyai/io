package testdata

import (
	"context"
	"fmt"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/dial/pipe"
	"github.com/injoyai/io/dial/proxy"
	"time"
)

func TestProxy() error {

	go proxy.SwapTCPClient(":10089", func(ctx context.Context, c *io.Client, e *proxy.Entity) {
		c.Debug()
		c.GoForWriter(time.Second*3, func(c *io.IWriter) error {
			e.Proxy(proxy.NewWriteMessage("key", "http://www.baidu.com", nil))
			return nil
		})
	})

	return proxy.SwapTCPServer(10089, func(s *io.Server) {
		s.Debug()
	})

	return nil
}

func ProxyClient(addr string) *io.Client {
	return proxy.SwapTCPClient(addr, func(ctx context.Context, c *io.Client, e *proxy.Entity) {
		c.Debug()
		//logs.Debug("重连...")
	})
}

func ProxyTransmit(port int) error {
	s, err := pipe.NewTransmit(dial.TCPListenFunc(port))
	if err != nil {
		return err
	}
	s.SetKey(fmt.Sprintf(":%d", port))
	s.Debug()
	return s.Run()
}

func VPNClient(serverPort int, clientAddr string) error {

	// 普通的tcpServer服务,用于监听用户数据
	s, err := proxy.NewServer(serverPort)
	if err != nil {
		return err
	}
	s.Debug()

	// 通道客户端,用于连接数据转发服务端,进行数据的封装
	// 所有数据都经过这个连接
	var c *io.Client
	go func() {
		c = pipe.RedialTCP(clientAddr, func(ctx context.Context, c *io.Client) {
			c.Debug()
			c.SetDealFunc(func(msg *io.IMessage) {
				s.Write(msg.Message)
			})
		})
	}()

	//设置数据处理函数
	s.SetDealFunc(func(msg *io.IMessage) error {
		if c == nil {
			return msg.Close()
		}
		_, err := c.Write(msg.Bytes())
		return err
	})

	//s.SetDealFunc(nil)

	//运行服务
	return s.Run()
}