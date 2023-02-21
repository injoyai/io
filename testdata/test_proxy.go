package testdata

import (
	"context"
	"fmt"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/dial/pipe"
	"github.com/injoyai/io/dial/proxy"
	"github.com/injoyai/logs"
	"strings"
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

					e.Resp(proxy.NewWriteMessage("key", "http://www.baidu.com", nil))

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
			logs.Debug(io.PrintfWithASCII(msg, append([]string{"P|C"}, tag...)...))
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

	var c *io.Client

	s, err := proxy.NewServer(serverPort, c)
	if err != nil {
		return err
	}
	s.Debug()

	go func() {
		c = pipe.Redial(dial.TCPFunc(clientAddr), func(ctx context.Context, c *io.Client) {
			c.Debug()
			c.SetPrintFunc(func(msg io.Message, tag ...string) {
				logs.Debug(io.PrintfWithASCII(msg, append([]string{"P|C"}, tag...)...))
			})
			c.SetWriteFunc(pipe.DefaultWriteFunc)
			c.SetDealFunc(func(msg *io.IMessage) {
				for _, v := range strings.Split(msg.String(), "}") {
					if len(v) > 0 {
						m, err := proxy.DecodeMessage([]byte(v + "}"))
						if err != nil {
							logs.Err(err)
							return
						}
						s.Write(m.GetData())
					}
				}
			})
		})
	}()

	return s.Run()
}
