package testdata

import (
	"testing"
)

func TestTestProxy(t *testing.T) {
	t.Log(TestProxy())
	select {}
}

func TestProxyTransmit(t *testing.T) {
	t.Log(ProxyTransmit(12000))
}

func TestProxyClient(t *testing.T) {
	ProxyClient(":12000")
	select {}
}

func TestVPNClient(t *testing.T) {
	t.Log(VPNClient(1082, 1090, ":12000"))
}
