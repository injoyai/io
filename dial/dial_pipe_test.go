package dial

import (
	"github.com/injoyai/io"
	"testing"
)

func TestNewTunnelServer(t *testing.T) {
	s, err := NewTCPServer(10088, func(s *io.Server) {
		s.Debug(true)
		s.SetPrintWithHEX()
	})
	if err != nil {
		t.Error(err)
		return
	}
	NewTunnelServer(s)
	t.Log(s.Run())
}

func TestNewTunnelClient(t *testing.T) {
	s, err := NewTCPServer(10086, func(s *io.Server) {
		s.Debug(true)
		s.SetPrintWithHEX()
	})
	if err != nil {
		t.Error(err)
		return
	}
	NewTunnelClient(s, WithTCP(":10088"), func(c *io.Client) {
		c.Debug()
		c.SetPrintWithASCII()
	})
	t.Log(s.Run())
}
