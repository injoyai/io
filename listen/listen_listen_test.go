package listen

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"testing"
	"time"
)

func TestTCPServer(t *testing.T) {
	s, err := NewTCPServer(10089)
	if err != nil {
		t.Error(err)
		return
	}
	s.Debug()
	s.SetPrintWithASCII()
	//s.SetDealFunc(func(msg *io.IMessage) {
	//	//msg.WriteString("HTTP/1.1 308 Moved Permanently\r\nLocation: http://www.baidu.com\r\n")
	//	msg.WriteString("HTTP/1.1 308 Moved Permanently\r\nLocation: /\r\n")
	//	msg.TryCloseWithDeadline()
	//})
	t.Error(s.Run())

}

func TestRedial(t *testing.T) {
	dial.RedialTCP(":10086", func(c *io.Client) {
		c.SetPrintWithASCII()
		c.Debug()
		//c.GoTimerWriter(time.Second*5, func(c *io.IWriter) error {
		//	_, err := c.WriteHEX("3a520600030a01000aaa0d")
		//	return err
		//})
	})
	select {}
}

func TestRunUDPServer(t *testing.T) {
	RunUDPServer(20001, func(s *io.Server) {
		s.Debug()
		s.SetPrintWithHEX()
		s.SetDealFunc(func(msg *io.IMessage) {
			msg.WriteString("7777")
		})
	})
}

// 测试传输速度
func TestIOSpeed(t *testing.T) {
	start := time.Now() //当前时间
	length := 20 << 20  //传输的数据大小
	go RunTCPServer(io.DefaultPort, func(s *io.Server) {
		s.Debug(false)
		s.SetReadWithAll()
		s.SetDealFunc(func(msg *io.IMessage) {
			t.Log("数据长度: ", msg.Len())
			t.Log("传输耗时: ", time.Now().Sub(start))
		})
	})
	<-dial.RedialTCP(io.DefaultPortStr, func(c *io.Client) {
		c.Debug(false)
		data := make([]byte, length)
		start = time.Now()
		c.Write(data)
		c.SetDealFunc(func(msg *io.IMessage) {
			//t.Log(msg)
		})
	}).DoneAll()
}
