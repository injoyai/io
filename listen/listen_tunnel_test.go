package listen

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"testing"
)

func TestNewTunnelServer(t *testing.T) {
	s, err := NewTCPServer(20088, func(s *io.Server) {
		s.Logger.Debug(true)
		s.Logger.SetPrintWithASCII()
	})
	if err != nil {
		t.Error(err)
		return
	}
	NewTunnelServer(s)
	t.Log(s.Run())
}

func TestNewTunnelClient(t *testing.T) {
	s, err := NewTCPServer(20086, func(s *io.Server) {
		s.Logger.Debug(true)
		s.Logger.SetPrintWithHEX()
	})
	if err != nil {
		t.Error(err)
		return
	}
	NewTunnelClient(s, dial.WithTCP(":20088"), "", func(c *io.Client) {
		c.Debug(true)
		c.SetPrintWithASCII()
	})
	t.Log(s.Run())
}
