package proxy

import (
	"testing"
)

func TestNew(t *testing.T) {
	e := New()
	_, err := e.Write(NewWriteMessage("", "www.baidu.com:80", []byte("GET /ping HTTP/1.1\r\n\r\n")).Bytes())
	if err != nil {
		t.Error(err)
		return
	}
}
