package proxy

import (
	"testing"
)

func TestNew(t *testing.T) {
	e := New()
	e.Debug()
	_, err := e.Write(NewWriteMessage("", "121.36.99.197:8000", []byte(`GET /ping HTTP/1.1
Host: 127.0.0.1
Connection: close

`)).Bytes())
	if err != nil {
		t.Error(err)
		return
	}
	select {}
}
