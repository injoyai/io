package testdata

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/listen"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	dial.RedialTCP(":10089", func(c *io.Client) {
		c.Debug()
		c.SetPrintWithUTF8()
		c.SetKey("test")
		c.GoTimerWriteString(time.Second*3, "666")
	})
	select {}
}

func TestNewServer(t *testing.T) {
	_, err := listen.NewTCPServer(10089, func(s *io.Server) {
		s.Debug()
		s.SetDealFunc(func(c *io.Client, msg io.Message) {
			c.WriteString("777")
		})
		t.Log(s.Run())
	})
	if err != nil {
		t.Error(err)
		return
	}
}

func TestCloseAll(t *testing.T) {
	t.Log(CloseAll(10089))
}

func TestTimeoutClient(t *testing.T) {
	t.Log(TimeoutClient(10089, time.Second*5))
}

func TestTimeoutServer(t *testing.T) {
	t.Log(TimeoutServer(10089, time.Second*5))
}

func TestGoFor(t *testing.T) {
	t.Log(GoFor(10089))
}

func TestServerMaxClient(t *testing.T) {
	t.Log(ServerMaxClient(10089))
}

func TestClientCtxParent(t *testing.T) {
	t.Log(ClientCtxParent(10089))
}

func TestPool(t *testing.T) {
	t.Log(Pool(10089))
}

//func TestPoolWrite(t *testing.T) {
//	p, err := PoolWrite(10089)
//	if err != nil {
//		t.Error(err)
//		return
//	}
//	for {
//		<-time.After(time.Second)
//		p.WriteString("666")
//	}
//}
