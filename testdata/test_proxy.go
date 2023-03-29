package testdata

import (
	"context"
	"errors"
	"fmt"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/dial/pipe"
	"github.com/injoyai/io/dial/proxy"
	"log"
	"net/http"
	"time"
)

func TestClient(addr string) {
	go func() {
		log.Println(http.ListenAndServe(fmt.Sprintf(":%d", 6067), nil))
	}()
	for range chans.Count(1000, time.Millisecond*10) {
		dial.RedialTCP(addr, func(ctx context.Context, c *io.Client) {
			c.Debug()
			c.GoTimerWriter(time.Second*3, func(c *io.IWriter) error {
				_, err := c.WriteString(time.Now().String())
				return err
			})
		})
	}
	select {}
}

func TestProxy() error {

	go proxy.SwapTCPClient(":10089", func(ctx context.Context, c *io.Client, e *proxy.Entity) {
		c.Debug()
		c.GoTimerWriter(time.Second*3, func(c *io.IWriter) error {
			e.AddMessage(proxy.NewWriteMessage("key", "http://www.baidu.com", nil))
			return nil
		})
	})

	return proxy.SwapTCPServer(10089, io.WithServerDebug())

	return nil
}

func ProxyClient(addr string) *io.Client {
	return proxy.SwapTCPClient(addr, proxy.WithClientDebug())
}

func ProxyTransmit(port int) error {
	s, err := pipe.NewTransmit(dial.TCPListenFunc(port))
	if err != nil {
		return err
	}
	s.SetKey(fmt.Sprintf(":%d", port))
	s.Debug()
	return s.Run()
}

func VPNClient(tcpPort, udpPort int, clientAddr string) error {

	// 普通的tcpServer服务,用于监听用户数据
	vpnClient, err := proxy.NewTCPServer(tcpPort, proxy.WithServerDebug(false))
	if err != nil {
		return err
	}

	// 通道客户端,用于连接数据转发服务端,进行数据的封装
	// 所有数据都经过这个连接
	var pipeClient *io.Client
	go pipe.RedialTCP(clientAddr, func(ctx context.Context, c *io.Client) {
		pipeClient = c
		c.Debug()
		c.SetDealFunc(proxy.DealWithServer(vpnClient))
	})

	//设置数据处理函数
	vpnClient.SetDealFunc(func(msg *proxy.Message) error {
		if pipeClient == nil {
			return errors.New("pipe未连接")
		}
		//发送到通道
		_, err := pipeClient.Write(msg.Bytes())
		return err
	})

	return vpnClient.Run()
}
