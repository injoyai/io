package dial

import (
	"context"
	"fmt"
	"github.com/injoyai/io"
	"testing"
	"time"
)

func TestNewWebsocket(t *testing.T) {
	url := "ws://192.168.10.24:10001/api/user/notice/ws"
	//"ws://192.168.10.3:1880/node-red/comms"
	url = "ws://192.168.10.24:10001/api/ai/info/runtime/ws?id=83"
	//url := "ws://192.168.10.24:10001/api/user/notice/ws"
	//url += "?token=jbYKl72cbOGvbVRwIqM4r6eoirw8f1JRD44+4D5E/URRY4L6TTZYYb/9yhedvd2Ii2GtLo9MieBy5FBeUhugK5jHvppFjExz3B5DVFPqsomF5wezKDFc8a2hZSQ9IDHTS/C+j/3ESSRdbkVHPFxbzQ=="
	//url = strings.ReplaceAll(url, "+", "%2B")
	t.Log(url)
	RedialWebsocket(url, nil, io.WithClientDebug(), func(ctx context.Context, c *io.Client) {
		c.GoTimerWriter(time.Second*10, func(w *io.IWriter) error {
			_, err := w.WriteString("83")
			return err
		})
	})
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

func TestRedialTCP(t *testing.T) {
	RedialTCP("192.168.10.24:10086", io.WithClientDebug(), func(ctx context.Context, c *io.Client) {
		c.WriteString("GET /sn/BFEBFBFF000906ED HTTP/1.1\r\n\r\n")
	})
	select {}
}

func TestRedialMQTT(t *testing.T) {
	//RedialMQTT("xxx")
}

func TestRedialSSH(t *testing.T) {
	RedialSSH(&SSHConfig{
		Addr:     "192.168.10.40:22",
		User:     "qinalang",
		Password: "ql1123",
	}, func(ctx context.Context, c *io.Client) {
		c.Debug()
		go func() {
			for {
				msg := ""
				fmt.Scan(&msg)
				c.WriteString(msg)
			}
		}()
	})
	select {}
}
