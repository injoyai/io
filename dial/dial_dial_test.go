package dial

import (
	"context"
	"github.com/injoyai/io"
	"testing"
	"time"
)

func TestNewWebsocket(t *testing.T) {
	//"ws://192.168.10.3:1880/node-red/comms"
	RedialWebsocket("ws://192.168.10.103:1880/comms", map[string][]string{
		"Sn": {"EA060900FFFBEBBF"},
	}, func(ctx context.Context, c *io.Client) {
		c.SetRedialMaxTime(time.Second * 2)
		c.Debug()
		c.SetDealQueueFunc(10, func(msg io.Message) {
			t.Log(msg)
		})
	})
	select {}
}
