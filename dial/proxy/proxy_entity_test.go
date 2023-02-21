package proxy

import (
	"testing"
)

func TestNew(t *testing.T) {
	e := New()
	e.Debug()
	_, err := e.Write(NewWriteMessage("", "192.168.10.40:10001", []byte("GET /ping HTTP/1.1\r\nHost 127.0.0.1\r\n\r\n")).Bytes())
	if err != nil {
		t.Error(err)
		return
	}
	select {}
}
