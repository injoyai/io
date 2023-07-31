package testdata

import (
	"testing"
)

func TestNewPortForwardingClient(t *testing.T) {
	NewPortForwardingClient(":10089")
	select {}
}

func TestNewPortForwardingServer(t *testing.T) {
	t.Error(NewPortForwardingServer(10089, 10000, "192.168.3.113:10001"))
}
