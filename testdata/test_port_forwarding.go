package testdata

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial/proxy"
	"github.com/injoyai/logs"
	"time"
)

// NewPortForwardingClient 端口转发客户端
func NewPortForwardingClient(addr string) error {
	return proxy.NewPortForwardingClient(addr, "sn", func(c *io.Client, e *proxy.Entity) {
		c.SetRedialMaxTime(time.Second * 3)
		c.SetPrintWithASCII()
		c.Debug()
		e.Debug()
	}).Run()
}

// NewPortForwardingServer 端口转发服务端
func NewPortForwardingServer(port int, proxyPort int, proxyAddr string) error {
	s, err := proxy.NewPortForwardingServer(port, func(s *io.Server) {
		s.Debug()
		s.SetPrintWithHEX()
	})
	if err != nil {
		return err
	}
	logs.PrintErr(s.Listen(proxyPort, "sn", proxyAddr))
	return s.Run()
}
