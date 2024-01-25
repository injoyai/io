package io

import (
	"encoding/hex"
	"testing"
)

func TestDecodeSimple(t *testing.T) {
	s := "68002001046b6579310476616c31046b6579320476616c3200070001020304050693"
	bs, _ := hex.DecodeString(s)
	p, err := DecodeSimple(bs)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(p)
	t.Log(p.Data.SMap())
}

func TestNewSimple(t *testing.T) {
	p := &Simple{
		Control: SimpleControl{
			Type: SimpleRead,
		},
		Data: SimpleData{
			"key1": []byte("val1"),
			"key2": []byte("val2"),
			"":     []byte{0, 1, 2, 3, 4, 5, 6},
		},
	}
	t.Log(p.Bytes().HEX())
	p2, err := DecodeSimple(p.Bytes())
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(p2.Bytes().HEX())
	t.Log(p2)
	t.Log(p2.Data)
}
