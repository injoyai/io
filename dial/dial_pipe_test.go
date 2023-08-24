package dial

import (
	"github.com/injoyai/io"
	"testing"
)

func TestNewTunnelServer(t *testing.T) {
	s, err := NewTunnelServer(TCPListenFunc(io.DefaultPort), func(s *io.Server) {
		s.Debug(true)
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(s.Run())
}
