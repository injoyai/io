package dial

import (
	"context"
	"github.com/injoyai/io"
	"testing"
	"time"
)

func TestTCPServer(t *testing.T) {
	s, err := io.NewServer(TCPListenFunc(10089))
	if err != nil {
		t.Error(err)
		return
	}
	s.Debug()
	s.SetPrintWithASCII()
	s.SetDealFunc(func(msg *io.IMessage) {
		//msg.WriteString("HTTP/1.1 308 Moved Permanently\r\nLocation: http://www.baidu.com\r\n")
		msg.WriteString("HTTP/1.1 308 Moved Permanently\r\nLocation: /\r\n")
		msg.TryCloseWithDeadline()
	})
	t.Error(s.Run())

}

func TestRedial(t *testing.T) {
	RedialTCP("121.36.99.197:10086", func(ctx context.Context, c *io.Client) {
		c.SetPrintWithASCII()
		c.Debug()
		c.GoTimerWriter(time.Second*5, func(c *io.IWriter) error {
			_, err := c.WriteHEX("3a520600030a01000aaa0d")
			return err
		})
	})
	select {}
}
