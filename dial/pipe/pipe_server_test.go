package pipe

import (
	"context"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"testing"
	"time"
)

func TestRedial(t *testing.T) {
	addr := ":10089"
	RedialTCP(addr, func(ctx context.Context, c *io.Client) {
		c.Debug()
		c.GoTimerWriter(time.Second*3, func(c *io.IWriter) error {
			_, err := c.WriteString("666")
			return err
		})
	})
	select {}
}

func TestNewServer(t *testing.T) {
	t.Error(func() error {
		s, err := NewServer(dial.TCPListenFunc(10089))
		if err != nil {
			return err
		}
		s.Debug()
		s.SetDealFunc(func(msg *io.IMessage) {
			msg.WriteString("777")
		})
		return s.Run()
	}())
}
