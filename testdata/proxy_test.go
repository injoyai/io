package testdata

import "testing"

func TestTestProxy(t *testing.T) {
	t.Log(TestProxy())
	select {}
}
