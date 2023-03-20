package rpc

import (
	"encoding/hex"
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
	p2, err := DecodePkg(bs)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%#v", p2)
	if p2.Type != p.Type || p2.MsgID != p.MsgID || hex.EncodeToString(p2.Data) != hex.EncodeToString(p.Data) {
		t.Error("解析错误")
	}
}
