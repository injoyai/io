package pipe

import (
	"context"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	c, err := NewClient(dial.TCPFunc("127.0.0.1:10089"))
	if err != nil {
		t.Error(err)
		return
	}
	c.Debug()
	c.Redial(func(ctx context.Context, c *io.Client) {
		c.GoForWriter(time.Second*3, func(c io.Writer) (int, error) {
			return c.Write([]byte("ping"))
		})
	})
	t.Error(c.Run())
	select {}
}
