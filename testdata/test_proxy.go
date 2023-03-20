package testdata

import (
	"context"
	"errors"
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
		c.GoTimerWriter(time.Second*3, func(c *io.IWriter) error {
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

func VPNClient(tcpPort, udpPort int, clientAddr string) error {

	// 普通的tcpServer服务,用于监听用户数据
	tcp, err := proxy.NewTCPServer(tcpPort, func(s *proxy.Server) { s.Debug() })
	if err != nil {
		return err
	}

	// 通道客户端,用于连接数据转发服务端,进行数据的封装
	// 所有数据都经过这个连接
	var pipeClient *io.Client
	go func() {
		pipeClient = pipe.RedialTCP(clientAddr, func(ctx context.Context, c *io.Client) {
			c.Debug()
			c.SetDealFunc(func(msg *io.IMessage) {
				m, err := proxy.DecodeMessage(msg.Message)
				if err == nil {
					switch m.ConnectType {
					default:
						//通道过来的数据
						tcp.WriteMessage(m)
					}
				}
			})
		})
	}()

	//设置数据处理函数
	tcp.Debug()
	tcp.SetDealFunc(func(msg *proxy.Message) error {
		if pipeClient == nil {
			return errors.New("pipe未连接")
		}
		_, err := pipeClient.Write(msg.Bytes())
		return err
	})

	return tcp.Run()
}
