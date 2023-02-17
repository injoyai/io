package testdata

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Log(NewClient(":10089"))
}

func TestNewServer(t *testing.T) {
	_, err := dial.NewTCPServer(10089, func(s *io.Server) {
		s.Debug()
		s.SetDealFunc(func(msg *io.ClientMessage) {
			msg.WriteString("777")
		})
		t.Log(s.Run())
	})
	if err != nil {
		t.Error(err)
		return
	}
}

func TestNewTestMustDialBug(t *testing.T) {
	t.Log(NewTestMustDialBug(10089))
}

func TestClientRun(t *testing.T) {
	ClientRun(":10089")
}
