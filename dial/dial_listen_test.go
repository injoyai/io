package dial

import (
	"context"
	"encoding/hex"
	"github.com/injoyai/io"
	"testing"
	"time"
)

func TestTCPServer(t *testing.T) {
	s, err := io.NewServer(func() (io.Listener, error) {
		return TCPListener(10089)
	})
	if err != nil {
		t.Error(err)
		return
	}
	s.Debug()
	s.SetDealFunc(func(msg *io.ClientMessage) {
		msg.WriteString("777")
	})
	t.Error(s.Run())

}

func TestRedial(t *testing.T) {
	io.Redial(TCPFunc("121.36.99.197:10086"), func(ctx context.Context, c *io.Client) {
		c.SetPrintWithASCII()
		c.Debug()
		c.GoForWriter(time.Second*5, func(c io.Writer) (int, error) {
			bs, _ := hex.DecodeString("3a520600030a01000aaa0d")
			return c.Write(bs)
		})
	})
	select {}
}
