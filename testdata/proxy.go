package testdata

import (
	"context"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial/proxy"
	"github.com/injoyai/logs"
	"time"
)

func TestProxy() error {

	go proxy.SwapTCPClient(":10089", func(ctx context.Context, c *io.Client, e *proxy.Entity) {
		c.SetPrintFunc(func(msg io.Message, tag ...string) {
			logs.Debug(io.PrintfWithASCII(msg, append([]string{"P|C"}, tag...)...))
		})
		c.Debug()
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
