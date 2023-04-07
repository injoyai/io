package dial

import (
	"context"
	"github.com/injoyai/io"
	"strings"
	"testing"
)

func TestNewWebsocket(t *testing.T) {
	//"ws://192.168.10.3:1880/node-red/comms"
	url := "ws://192.168.10.24:8200/ops/notice/ws"
	url += "?token=jbYKl72cbOGvbVRwIqM4r6eoirw8f1JRD44+4D5E/URRY4L6TTZYYb/9yhedvd2Ii2GtLo9MieBy5FBeUhugK5jHvppFjExz3B5DVFPqsomF5wezKDFc8a2hZSQ9IDHTS/C+j/3ESSRdbkVHPFxbzQ=="
	url = strings.ReplaceAll(url, "+", "%2B")
	t.Log(url)
	RedialWebsocket(url, map[string][]string{}, io.WithClientDebug())
	select {}
}

func TestNewTCP(t *testing.T) {
	//"ws://192.168.10.3:1880/node-red/comms"
	RedialTCP(":1082", func(ctx context.Context, c *io.Client) {
		c.Debug()
		c.WriteAny("666")
	})
	select {}
}

func TestRtsp(t *testing.T) {
	RedialTCP("34.227.104.115:554", io.WithClientDebug())
	select {}
}
