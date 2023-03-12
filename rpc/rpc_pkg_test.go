package rpc

import (
	"testing"
)

func TestDecodePkg(t *testing.T) {
	p := &Pkg{
		Type:  0,
		MsgID: 1,
		Data:  []byte{2, 3},
	}
	bs := p.Bytes()
	t.Log(bs.HEX())
	p, err := DecodePkg(bs)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%#v", p)
}
