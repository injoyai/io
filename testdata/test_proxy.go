package testdata

import (
	"fmt"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"log"
	"net/http"
	"time"
)

func TestClient(addr string) {
	go func() {
		log.Println(http.ListenAndServe(fmt.Sprintf(":%d", 6067), nil))
	}()
	for range chans.Count(1000, time.Millisecond*10) {
		dial.RedialTCP(addr, func(c *io.Client) {
			c.Debug()
			c.GoTimerWriteString(time.Second*3, time.Now().String())
		})
	}
	select {}
}

//func TestProxy() error {
//
//	go proxy.NewTCPClient(":10089", func(c *io.Client, e *proxy.Entity) {
//		c.Debug()
//		c.GoTimerWriter(time.Second*3, func(c *io.Client) error {
//			e.AddMessage(proxy.NewWriteMessage("key", "http://www.baidu.com", nil))
//			return nil
//		})
//	})
//
//	return proxy.NewSwapTCPServer(10089, func(s *io.Server) {
//		s.Debug()
//	})
//
//	return nil
//}

//func ProxyClient(addr string) *io.Client {
//	return proxy.NewTCPClient(addr)
//}

//func ProxyTransmit(port int) error {
//	s, err := dial.NewPipeTransmit(port)
//	if err != nil {
//		return err
//	}
//	s.SetKey(fmt.Sprintf(":%d", port))
//	//s.Debug()
//	return s.Run()
//}

//func VPNClient(tcpPort, udpPort int, clientAddr string) error {
//
//	// 普通的tcpServer服务,用于监听用户数据
//	vpnClient, err := proxy.NewTCPServer(tcpPort)
//	if err != nil {
//		return err
//	}
//
//	// 通道客户端,用于连接数据转发服务端,进行数据的封装
//	// 所有数据都经过这个连接
//	var pipeClient *io.Client
//	go dial.RedialPipe(clientAddr, func(c *io.Client) {
//		pipeClient = c
//		//c.Debug()
//		c.SetDealWithWriter(vpnClient)
//	})
//
//	//设置数据处理函数
//	vpnClient.Debug()
//	vpnClient.SetDealFunc(func(msg *proxy.CMessage) error {
//		if pipeClient == nil {
//			return errors.New("pipe未连接")
//		}
//		//发送到通道
//		_, err := pipeClient.Write(msg.Bytes())
//		return err
//	})
//
//	return vpnClient.Run()
//}
