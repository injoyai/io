package dial

import (
	"github.com/injoyai/io"
	"testing"
)

func TestTCPServer(t *testing.T) {
	s, err := io.NewServer(func() (io.Listener, error) {
		return TCPListener(10089)
	})
	if err != nil {
		t.Error(err)
		return
	}
	s.Debug()
	s.SetDealFunc(func(msg *io.ClientMessage) {
		msg.WriteString("777")
	})
	t.Error(s.Run())

}
