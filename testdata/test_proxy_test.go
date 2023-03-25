package testdata

import (
	"github.com/injoyai/base/chans"
	"github.com/injoyai/base/maps/wait"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"testing"
	"time"
)

func TestTestProxy(t *testing.T) {
	t.Log(TestProxy())
	select {}
}

func TestProxyTransmit(t *testing.T) {
	t.Log(ProxyTransmit(12000))
}

func TestProxyClient(t *testing.T) {
	ProxyClient(":12000")
	select {}
}

func TestVPNClient(t *testing.T) {
	t.Log(VPNClient(1082, 1090, ":12000"))
}

func TestVPNHTTP(t *testing.T) {

	for range chans.Count(1000) {
		<-time.After(time.Second * 3)
		c, err := dial.NewTCP(":1082")
		if err != nil {
			t.Error(err)
			continue
		}

		c.Debug()
		c.SetDealFunc(func(msg *io.IMessage) {
			wait.Done("", nil)
		})
		go c.Run()
		c.WriteString(`CONNECT / HTTP/1.1
Host: 121.36.99.197:8000

`)
		go func() {
			if _, err := wait.Wait(""); err == nil {
				c.WriteString(`GET /ping HTTP/1.1
Host: 121.36.99.197:8000

				`)
			}
		}()
	}
	select {}

}
