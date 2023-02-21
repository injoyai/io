package testdata

import (
	"context"
	"fmt"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/dial/pipe"
	"github.com/injoyai/io/dial/proxy"
	"github.com/injoyai/logs"
	"net/http"
	"net/url"
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

	s, err := io.NewServer(dial.TCPListenFunc(serverPort))
	if err != nil {
		return err
	}
	s.SetKey(fmt.Sprintf(":%d", serverPort))
	s.Debug()
	s.SetPrintFunc(func(msg io.Message, tag ...string) {
		logs.Debug(io.PrintfWithASCII(msg, append([]string{"P|S"}, tag...)...))
	})

	var c *io.Client
	go func() {
		c = pipe.Redial(dial.TCPFunc(clientAddr), func(ctx context.Context, c *io.Client) {
			c.Debug()
			c.SetPrintFunc(func(msg io.Message, tag ...string) {
				logs.Debug(io.PrintfWithASCII(msg, append([]string{"P|C"}, tag...)...))
			})
			c.SetDealFunc(func(msg *io.IMessage) {
				for _, v := range strings.Split(msg.String(), "}") {
					if len(v) > 0 {
						m, err := proxy.DecodeMessage([]byte(v + "}"))
						if err != nil {
							logs.Err(err)
							return
						}
						logs.Debug(string(m.GetData()))
						s.Write(m.GetData())
					}
				}

			})
			//c.GoForWriter(time.Second*5, func(c *io.IWriter) (int, error) {
			//	return c.Write(proxy.NewWriteMessage("test", "192.168.10.40:10001", []byte("GET /ping HTTP/1.1 \r\nHost 192.168.10.40:10001\r\n")).Bytes())
			//})
		})
	}()

	s.SetDealFunc(func(msg *io.IMessage) {
		if c == nil {
			return
		}
		// HTTP 请求
		if list := strings.Split(msg.String(), " "); len(list) > 2 && strings.Contains(list[2], "HTTP") {
			if list[0] == http.MethodConnect {
				// HTTP 代理请求
				msg.Client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
				return
			} else {
				// HTTP 普通请求
				u, err := url.Parse(list[1])
				if err == nil {
					port := u.Port()
					if len(port) == 0 {
						switch strings.ToLower(u.Scheme) {
						case "https":
							port = "443"
						default:
							port = "80"
						}
					}
					addr := fmt.Sprintf("%s:%s", u.Hostname(), port)
					c.WriteAny(proxy.NewWriteMessage("test", addr, msg.Bytes()))
				}
			}
		}
	})
	return s.Run()
}
