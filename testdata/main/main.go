package main

import (
	"context"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/cfg"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial/proxy"
	"github.com/injoyai/io/testdata"
	"github.com/injoyai/logs"
)

func main() {
	//logs.PrintErr(NewPortForwardingServer())
	//return
	NewPortForwardingClient()
	return
	testdata.VPNClient(1082, 1090, ":12000")
	testdata.ProxyTransmit(12000)
	testdata.ProxyClient(":12000")
	select {}
}

func NewPortForwardingClient() {
	serverAddr := ""
	for len(serverAddr) == 0 {
		fmt.Println("请输入服务地址(默认121.36.99.197:9000):")
		fmt.Scanln(&serverAddr)
	}
	sn := ""
	fmt.Println("请输入SN(默认test):")
	fmt.Scanln(&sn)
	if len(sn) == 0 {
		sn = "test"
	}
	proxyAddr := ""
	fmt.Println("请输入代理地址(默认代理全部):")
	fmt.Scanln(&proxyAddr)
	c := proxy.NewPortForwardingClient(serverAddr, sn, func(ctx context.Context, c *io.Client, e *proxy.Entity) {
		c.Debug()
		if len(proxyAddr) > 0 {
			e.SetWriteFunc(func(msg *proxy.Message) (*proxy.Message, error) {
				msg.Addr = proxyAddr
				return msg, nil
			})
		}
	})
	c.Run()
	select {}
}

func NewPortForwardingServer() error {

	port := cfg.GetInt("port", 9000)
	s, err := proxy.NewPortForwardingServer(port)
	if err != nil {
		return err
	}
	s.Debug()
	for _, v := range cfg.GetStrings("listen") {
		m := conv.NewMap(v)
		logs.PrintErr(s.Listen(m.GetInt("port"), m.GetString("sn"), m.GetString("addr")))
	}

	return s.Run()
}
