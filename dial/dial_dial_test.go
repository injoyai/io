package dial

import (
	"context"
	"github.com/injoyai/io"
	"testing"
)

func TestNewWebsocket(t *testing.T) {
	//"ws://192.168.10.3:1880/node-red/comms"
	RedialWebsocket("ws://192.168.10.24:8200/api/cache/log/ws", map[string][]string{
		"Sec-WebSocket-Protocol": {"BFEBFBFF000906ED"},
	}, func(ctx context.Context, c *io.Client) {
		c.Debug()
	})
	select {}
}
