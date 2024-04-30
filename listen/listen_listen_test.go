package listen

import (
	"errors"
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
	s.SetPrintWithUTF8()
	s.SetLevel(io.LevelAll)
	s.SetDealFunc(func(c *io.Client, msg io.Message) {
		//msg.WriteString("HTTP/1.1 308 Moved Permanently\r\nLocation: http://www.baidu.com\r\n")
		c.WriteString("HTTP/1.1 308 Moved Permanently\r\nLocation: /\r\n")
		//msg.TryCloseWithDeadline()
	})
	t.Error(s.Run())

}

func TestRedial(t *testing.T) {
	dial.RedialTCP(":10086", func(c *io.Client) {
		c.SetPrintWithUTF8()
		c.Debug()
		c.GoTimerWriteHEX(time.Second*5, "3a520600030a01000aaa0d")
	})
	select {}
}

func TestRunUDPServer(t *testing.T) {
	RunUDPServer(20001, func(s *io.Server) {
		s.Debug()
		s.SetPrintWithHEX()
		s.SetDealFunc(func(c *io.Client, msg io.Message) {
			c.WriteString("7777")
		})
	})
}

// 测试传输速度
func TestIOSpeed(t *testing.T) {
	start := time.Now() //当前时间
	length := 20 << 20  //传输的数据大小
	go RunTCPServer(io.DefaultPort, func(s *io.Server) {
		s.SetLevel(io.LevelInfo)
		s.Debug(false)
		s.SetReadWith1KB() //100毫秒
		s.SetReadWithMB(1) //65毫秒
		s.SetDealFunc(func(c *io.Client, msg io.Message) {
			t.Log("数据长度: ", msg.Len())
			t.Log("传输耗时: ", time.Now().Sub(start))
		})
	})
	<-dial.RedialTCP(io.DefaultPortStr, func(c *io.Client) {
		c.Debug(false)
		data := make([]byte, length)
		start = time.Now()
		c.Write(data)
		c.SetDealFunc(func(c *io.Client, msg io.Message) {
			//t.Log(msg)
		})
	}).DoneAll()
}

func TestServerErr(t *testing.T) {
	s, err := NewTCPServer(1)
	if err != nil {
		t.Error(err)
		return
	}
	go func() {
		t.Error(s.Run())
		t.Error(s.Err())
	}()
	go func() {
		<-time.After(time.Second * 5)
		s.CloseWithErr(errors.New("测试错误"))
	}()
	t.Error(s.Err())
	select {}
}
