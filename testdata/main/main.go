package main

import (
	"bufio"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/cfg"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/dial/proxy"
	"github.com/injoyai/io/testdata"
	"github.com/injoyai/logs"
	"os"
	"runtime"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	dial.RedialWebsocket("ws://192.168.10.24:8300/v1/chat/ws", nil, func(c *io.Client) {
		c.Debug()
		c.SetPrintFunc(func(msg io.Message, tag ...string) {
			if len(tag) > 0 {
				switch tag[0] {
				case io.TagRead:
					fmt.Printf("[客服] %s\n", conv.NewMap(msg).GetString("data"))
				case io.TagWrite:
				default:
					fmt.Printf("[%s] %s\n", tag[0], msg)
				}
			}
		})
		go func() {
			for {
				msg, _ := reader.ReadString('\n')
				c.WriteString(msg)
			}
		}()
	})
	select {}

	logs.PrintErr(NewPortForwardingServer())
	return
	NewPortForwardingClient()
	return
	testdata.VPNClient(1082, 1090, ":12000")
	testdata.ProxyTransmit(12000)
	testdata.ProxyClient(":12000")
	select {}
}

func NewPortForwardingClient() {

	//服务端地址
	serverAddr := cfg.GetString("addr")
	if runtime.GOOS == "windows" && len(serverAddr) == 0 {
		fmt.Println("请输入服务地址(默认121.36.99.197:9000):")
		fmt.Scanln(&serverAddr)
		if len(serverAddr) == 0 {
			serverAddr = "121.36.99.197:9000"
		}
	}

	//客户端唯一标识
	sn := cfg.GetString("sn")
	if runtime.GOOS == "windows" && len(sn) == 0 {
		fmt.Println("请输入SN(默认test):")
		fmt.Scanln(&sn)
		if len(sn) == 0 {
			sn = "test"
		}
	}

	//代理地址
	proxyAddr := ""
	if runtime.GOOS == "windows" {
		fmt.Println("请输入代理地址(默认代理全部):")
		fmt.Scanln(&proxyAddr)
	}

	c := proxy.NewPortForwardingClient(serverAddr, sn, func(c *io.Client, e *proxy.Entity) {
		c.SetPrintWithBase()
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
	s, err := proxy.NewPortForwardingServer(port, func(s *io.Server) {
		s.SetPrintWithBase()
		s.Debug()
	})
	if err != nil {
		return err
	}
	for _, v := range cfg.GetStrings("listen") {
		m := conv.NewMap(v)
		logs.PrintErr(s.Listen(m.GetInt("port"), m.GetString("sn"), m.GetString("addr")))
	}

	return s.Run()
}
