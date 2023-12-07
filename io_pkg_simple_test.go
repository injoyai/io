package io

import (
	"encoding/hex"
	"testing"
)

func TestDecodeSimple(t *testing.T) {
	s := "68000f000106e6b58be8af95043636363634"
	bs, _ := hex.DecodeString(s)
	p, err := DecodeSimple(bs)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(p)
}
