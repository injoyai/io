package testdata

import "testing"

func TestNewClient(t *testing.T) {
	t.Log(NewClient(":10089"))
}

func TestNewServer(t *testing.T) {
	t.Log(NewServer(10089))
}

func TestNewTestMustDialBug(t *testing.T) {
	t.Log(NewTestMustDialBug(10089))
}
