package io

import (
	"bytes"
	"testing"
	"time"
)

func TestSwap(t *testing.T) {
	r1 := NewReadWriter(Read(func(p []byte) (int, error) {
		<-time.After(time.Second)
		return copy(p, []byte("r1")), nil
	}), Write(func(p []byte) (int, error) {
		t.Log(string(p))
		return len(p), nil
	}))
	r2 := NewReadWriter(Read(func(p []byte) (int, error) {
		<-time.After(time.Second)
		return copy(p, []byte("r2")), nil
	}), Write(func(p []byte) (int, error) {
		t.Log(string(p))
		return len(p), nil
	}))
	Swap(r1, r2)
}

func TestReadPrefix(t *testing.T) {
	r := bytes.NewReader([]byte("hello world woworld"))
	t.Log(ReadPrefix(r, []byte("llo"))) //nil
	t.Log(ReadPrefix(r, []byte("wor"))) //nil
	t.Log(ReadPrefix(r, []byte("wor"))) //nil
	t.Log(ReadPrefix(r, []byte("llo"))) //EOF
	t.Log(ReadPrefix(r, []byte("aaa"))) //EOF
	t.Log(ReadPrefix(r, []byte("aaa"))) //EOF
}
