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
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				<-time.After(time.Second * 3)
				if _, err := c.WriteAny(&Message{Data: []byte("ping")}); err != nil {

				}
			}
		}(c.Ctx())
	})
	t.Error(c.Run())
	select {}
}
