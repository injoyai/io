package testdata

import (
	"fmt"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
	"time"
)

func Example() {
	io.Redial(dial.WithTCP("xxx"))
	io.Redial(dial.WithUDP("xxx"))
	io.Redial(dial.WithSerial(nil))
	io.Redial(dial.WithFile("./xxx.txt"))

}

func NewClient(addr string) error {
	c := io.Redial(dial.WithTCP(addr), func(c *io.Client) {
		logs.Debug("重连...")
		c.Debug()
	})
	time.Sleep(time.Second * 5)
	logs.Debug("主动关闭")
	c.Close()
	select {}
	return nil
}

func NewServer(port int) error {
	s, err := io.NewServer(listen.WithTCP(port))
	if err != nil {
		return err
	}
	s.Debug()
	return s.Run()
}

// CloseAll 测试closeAll
func CloseAll(port int) error {

	s, err := listen.NewTCPServer(port, func(s *io.Server) {
		s.Debug()
	})
	if err != nil {
		return err
	}
	go func() {
		logs.Err("服务端:", s.Run())
	}()

	c := dial.RedialTCP(fmt.Sprintf(":%d", port), func(c *io.Client) {
		c.Debug()
	})

	<-time.After(time.Second * 1)
	logs.Debug("关闭服务端")
	s.Close()

	<-time.After(time.Second * 5)
	logs.Debug("关闭客户端")
	c.Close()

	<-time.After(time.Second * 10)
	logs.Debug("关闭客户端2")
	c.CloseAll()

	select {}
	return nil
}

func ClientRun(addr string) *io.Client {
	return io.Redial(dial.WithTCP(addr), func(c *io.Client) {
		c.Debug()
		c.SetPrintWithUTF8()
		c.SetKey("test")
		c.SetDealFunc(func(c *io.Client, msg io.Message) {
			logs.Debug(msg.String())
			//业务逻辑
			c.WriteString("666")
		})
	})
}

// TimeoutClient 测试客户端超时
func TimeoutClient(port int, timeout time.Duration) error {
	go dial.RedialTCP(fmt.Sprintf(":%d", port),
		func(c *io.Client) {
			c.Debug()
			c.SetReadTimeout(timeout)
		})
	s, err := io.NewServer(listen.WithTCP(port))
	if err != nil {
		return err
	}
	s.Debug()
	return s.Run()
}

// TimeoutServer 测试服务端超时
func TimeoutServer(port int, timeout time.Duration) error {
	go io.Redial(dial.WithTCP(fmt.Sprintf(":%d", port)),
		func(c *io.Client) {
			c.Debug()
		})
	s, err := io.NewServer(listen.WithTCP(port))
	if err != nil {
		return err
	}
	s.Debug()
	s.SetTimeout(timeout)
	return s.Run()
}

// GoFor 测试客户端的GoFor函数
func GoFor(port int) error {
	s, err := listen.NewTCPServer(port, func(s *io.Server) {
		s.Debug()
	})
	if err != nil {
		return err
	}
	c := dial.RedialTCP(fmt.Sprintf(":%d", port), func(c *io.Client) {
		c.Debug()
		c.GoTimerWriteString(time.Second*3, "666")
	})
	_ = c
	//c.GoTimer(time.Second*5, func(c *io.Client) error {
	//	logs.Debug(777)
	//	c.Close()
	//	return nil
	//})
	return s.Run()
}

// ServerMaxClient 测试服务端最大连接数
func ServerMaxClient(port int) error {
	s, err := listen.NewTCPServer(port, func(s *io.Server) {
		s.Debug()
		s.SetMaxClient(1)
	})
	if err != nil {
		return err
	}
	for i := 0; i < 2; i++ {
		dial.RedialTCP(fmt.Sprintf(":%d", port), func(c *io.Client) {
			c.Debug()
			c.SetKeepAlive(time.Second)
		})
	}
	return s.Run()
}

// ClientCtxParent 测试ctxAll
func ClientCtxParent(port int) error {
	s, err := io.NewServer(listen.WithTCP(port))
	if err != nil {
		return err
	}
	<-time.After(time.Second * 5)
	go s.Run()
	c := io.Redial(dial.WithTCP(fmt.Sprintf(":%d", port)),
		func(c *io.Client) {
			c.Debug()
			c.GoTimerWriteString(time.Second, "666")
		})
	//c.GoTimer(time.Second, func(c *io.Client) error {
	//	logs.Debug(777)
	//	return nil
	//})
	go func() {
		<-time.After(time.Second * 5)
		//断开客户端
		logs.Debug("断开客户端")
		s.CloseClientAll()
		//等待客户端重连
		<-time.After(time.Second * 5)
		logs.Debug("关闭客户端")
		c.CloseAll()
	}()
	select {}
	return nil
}

func Pool(port int) error {
	s, err := listen.NewTCPServer(port, func(s *io.Server) {
		s.Debug()
	})
	if err != nil {
		return err
	}
	io.NewPool(dial.WithTCP(fmt.Sprintf(":%d", port)), func(c *io.Client) {
		c.Debug()
		c.GoTimerWriteString(time.Second*10, "666")
	})
	return s.Run()
}

func PoolWrite(port int) (*io.Pool, error) {
	s, err := listen.NewTCPServer(port, func(s *io.Server) {
		s.Debug()
	})
	if err != nil {
		return nil, err
	}
	p := io.NewPool(dial.WithTCP(fmt.Sprintf(":%d", port)), func(c *io.Client) {
		c.Debug()
		c.GoTimerWriteString(time.Second*10, "666")
	})
	go s.Run()
	return p, nil
}
