package testdata

import (
	"context"
	"fmt"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/dial/pipe"
	"github.com/injoyai/io/dial/proxy"
	"github.com/injoyai/logs"
	"time"
)

func TestProxy() error {

	go proxy.SwapTCPClient(":10089", func(ctx context.Context, c *io.Client, e *proxy.Entity) {
		c.SetPrintFunc(func(msg io.Message, tag ...string) {
			logs.Debug(io.PrintfWithASCII(msg, append([]string{"P|C"}, tag...)...))
		})
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
				case <-time.After(time.Second * 3):

					e.Proxy(proxy.NewWriteMessage("key", "http://www.baidu.com", nil))

				}

			}
		}(ctx)
	})

	return proxy.SwapTCPServer(10089, func(s *io.Server) {
		s.SetPrintFunc(func(msg io.Message, tag ...string) {
			logs.Debug(io.PrintfWithASCII(msg, append([]string{"P|S"}, tag...)...))
		})
	})

}

func ProxyClient(addr string) *io.Client {
	return proxy.SwapTCPClient(addr, func(ctx context.Context, c *io.Client, e *proxy.Entity) {
		c.Debug()
		c.SetPrintFunc(func(msg io.Message, tag ...string) {
			if len(tag) > 0 && len(msg) > 4 {
				switch tag[0] {
				case io.TagWrite:
					bs, err := pipe.DefaultDecode(msg)
					if err != nil {
						logs.Err(err)
						return
					}
					m, err := proxy.DecodeMessage(bs)
					if err != nil {
						logs.Err(err)
						return
					}
					logs.Debugf("[PI|C][发送] %s", m.String())
				case io.TagRead:
					logs.Debug(msg.String())
					m, err := proxy.DecodeMessage(msg)
					if err != nil {
						logs.Err(err)
						return
					}
					logs.Debugf("[PI|C][接收] %s", m.String())
				default:
					logs.Debug(io.PrintfWithASCII(msg, append([]string{"PI|C"}, tag...)...))
				}
			}

			//logs.Debug(io.PrintfWithASCII(msg, append([]string{"PI|C"}, tag...)...))
		})
	})
}

func ProxyTransmit(port int) error {
	s, err := pipe.NewTransmit(dial.TCPListenFunc(port))
	if err != nil {
		return err
	}
	s.SetKey(fmt.Sprintf(":%d", port))
	s.Debug()
	s.SetPrintFunc(func(msg io.Message, tag ...string) {
		logs.Debug(io.PrintfWithASCII(msg, append([]string{"P|T"}, tag...)...))
	})
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
			//c.Debug()
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
