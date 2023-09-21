package dial

import (
	"fmt"
	"github.com/injoyai/base/g"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/io"
	"github.com/injoyai/logs"
	"testing"
	"time"
)

func TestRedialWebsocket(t *testing.T) {
	url := "ws://127.0.0.1/api/user/notice/ws"
	//"ws://192.168.10.3:1880/node-red/comms"
	//url = "ws://192.168.10.24:10001/api/ai/info/runtime/ws?id=83"
	//url = "ws://192.168.10.38:80/api/ai/photo/ws?key=0.0"
	//url := "ws://192.168.10.24:10001/api/user/notice/ws"
	//url += "?token=jbYKl72cbOGvbVRwIqM4r6eoirw8f1JRD44+4D5E/URRY4L6TTZYYb/9yhedvd2Ii2GtLo9MieBy5FBeUhugK5jHvppFjExz3B5DVFPqsomF5wezKDFc8a2hZSQ9IDHTS/C+j/3ESSRdbkVHPFxbzQ=="
	//url = strings.ReplaceAll(url, "+", "%2B")
	logs.Debug(url)
	c := RedialWebsocket(url, nil, func(c *io.Client) {
		c.Debug()
		//c.GoTimerWriteASCII(time.Second, "666")
	})
	<-c.DoneAll()
}

func TestRedialWebsocket2(t *testing.T) {
	start := time.Now()
	count := safe.NewInt32(0)
	go func() {
		for range g.Interval(time.Second) {
			sec := time.Now().Sub(start).Seconds()
			t.Log(int(float64(count.Get())/sec), "帧")
		}
	}()

	url := "ws://192.168.10.40:10001/api/ai/photo/v2/ws?key=150.0&Sn=null"
	url = "ws://192.168.10.15:80/api/ai/photo/v2/ws?key=0.0&Sn=null"
	c := RedialWebsocket(url, nil, func(c *io.Client) {
		c.Debug(false)
		c.SetDealFunc(func(msg *io.IMessage) {
			count.Add(1)
		})
	})
	<-c.DoneAll()
}

func TestRedialTCP(t *testing.T) {
	//"ws://192.168.10.3:1880/node-red/comms"
	RedialTCP(":10089", func(c *io.Client) {
		c.Debug()
		c.GoTimerWriteBytes(time.Second, []byte("666"))
	})
	select {}
}

func TestRedialRtsp(t *testing.T) {
	RedialTCP("34.227.104.115:554", func(c *io.Client) {
		c.Debug()
	})
	select {}
}

func TestRedialHTTP(t *testing.T) {
	RedialTCP("192.168.10.24:10086", func(c *io.Client) {
		c.Debug()
		c.WriteString("GET /sn/BFEBFBFF000906ED HTTP/1.1\r\n\r\n")
	})
	select {}
}

func TestRedialMQTT(t *testing.T) {
	//RedialMQTT("xxx")
}

func TestRedialSSH(t *testing.T) {
	RedialSSH(&SSHConfig{
		Addr:     "injoy:10022",
		User:     "root",
		Password: "root",
	}, func(c *io.Client) {
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

func TestRedialUDP(t *testing.T) {
	RedialUDP("127.0.0.1")
}

func TestRedialTCP2(t *testing.T) {
	RedialUDP("127.0.0.1")
}
