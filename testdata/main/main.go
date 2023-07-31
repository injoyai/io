package main

import (
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/cfg"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial/proxy"
	"github.com/injoyai/io/testdata"
	"github.com/injoyai/logs"
	"runtime"
)

func main() {
	logs.Err(NewPortForwardingServer())
	return
	NewPortForwardingClient()
	return
	testdata.VPNClient(1082, 1090, ":12000")
	testdata.ProxyTransmit(12000)
	testdata.ProxyClient(":12000")
	select {}
}

func Test() {
	//url := "ws://127.0.0.1:10001/api/user/notice/ws"
	//"ws://192.168.10.3:1880/node-red/comms"
	//url = "ws://192.168.10.24:10001/api/ai/info/runtime/ws?id=83"
	//url = "ws://192.168.10.38:80/api/ai/photo/ws?key=0.0"
	//url := "ws://192.168.10.24:10001/api/user/notice/ws"
	//url += "?token=jbYKl72cbOGvbVRwIqM4r6eoirw8f1JRD44+4D5E/URRY4L6TTZYYb/9yhedvd2Ii2GtLo9MieBy5FBeUhugK5jHvppFjExz3B5DVFPqsomF5wezKDFc8a2hZSQ9IDHTS/C+j/3ESSRdbkVHPFxbzQ=="
	//url = strings.ReplaceAll(url, "+", "%2B")
	//logs.Debug(url)
	//c := dial.RedialWebsocket(url, nil, func(c *io.Client) {
	//	c.Debug()
	//	//c.GoTimerWriteASCII(time.Second, "666")
	//})
	//<-c.DoneAll()
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
