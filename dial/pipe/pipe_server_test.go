package pipe

import (
	"github.com/injoyai/io/dial"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	s, err := NewServer(dial.TCPListenFunc(10089))
	if err != nil {
		t.Error(err)
		return
	}
	s.Debug()
	go func() {
		for {
			<-time.After(time.Second * 3)
			//s.WriteClientAll(conv.Bytes(&Message{Data: []byte("pong")}))
		}
	}()
	t.Error(s.Run())
}
