package testdata

import (
	"github.com/injoyai/io/dial/proxy"
	"github.com/injoyai/logs"
)

// NewPortForwardingClient 端口转发客户端
func NewPortForwardingClient(addr string) error {
	return proxy.NewPortForwardingClient(addr, "sn", proxy.WithClientDebug()).Run()
}

// NewPortForwardingServer 端口转发服务端
func NewPortForwardingServer(port int) error {
	s, err := proxy.NewPortForwardingServer(port)
	if err != nil {
		return err
	}
	s.Debug()
	logs.PrintErr(s.Listen(10000, "sn", "192.168.10.24:10001"))
	return s.Run()
}
