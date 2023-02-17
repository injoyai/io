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
	s, err := io.NewServer(dial.TCPListenFunc(10089))
	if err != nil {
		t.Error(err)
		return
	}
	s.Debug()
	s.SetDealFunc(func(msg *io.ClientMessage) {
		msg.WriteString("777")
	})
	t.Log(s.Run())
}

func TestNewTestMustDialBug(t *testing.T) {
	t.Log(NewTestMustDialBug(10089))
}

func TestClientRun(t *testing.T) {
	ClientRun(":10089")
}
