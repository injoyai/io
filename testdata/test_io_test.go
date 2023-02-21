package testdata

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Log(NewClient(":10089"))
}

func TestNewServer(t *testing.T) {
	_, err := dial.NewTCPServer(10089, func(s *io.Server) {
		s.Debug()
		s.SetDealFunc(func(msg *io.IMessage) {
			msg.WriteString("777")
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

func TestClientRun(t *testing.T) {
	ClientRun(":10089")

}

func TestTimeoutClient(t *testing.T) {
	t.Log(TimeoutClient(10089, time.Second*5))
}

func TestTimeoutServer(t *testing.T) {
	t.Log(TimeoutServer(10089, time.Second*5))
}
