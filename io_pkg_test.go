package io

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"testing"
)

func TestNewPkg(t *testing.T) {

	data := []byte{0, 1, 2, 3, 5}
	{
		t.Log(NewPkg(20, data).Bytes().HEX())

		s := "888800000014000014000102030512EC39DE8989"
		bs, err := hex.DecodeString(s)
		if err != nil {
			t.Error(err)
			return
		}

		p, err := DecodePkg(bs)
		if err != nil {
			t.Error(err)
			return
		}

		t.Logf("%#v", p)

		if hex.EncodeToString(p.Data) != hex.EncodeToString(data) {
			t.Errorf("解析失败,预期(%x),得到(%x)", data, p.Data)
		}
	}
	{
		t.Log(NewPkg(20, data).SetCompress(1).Bytes().HEX())

		s := "88880000002C0800141F8B08000000000000FF626064626605040000FFFF5AE35D2605000000A456D7048989"
		bs, err := hex.DecodeString(s)
		if err != nil {
			t.Error(err)
			return
		}

		p, err := DecodePkg(bs)
		if err != nil {
			t.Error(err)
			return
		}

		t.Logf("%#v", p)

		if hex.EncodeToString(p.Data) != hex.EncodeToString(data) {
			t.Errorf("解析失败,预期(%x),得到(%x)", data, p.Data)
		}
	}
	{

		for i := 0; i < 10000; i++ {
			data = append(data, byte(i))
		}
		bs := NewPkg(20, data).Bytes()
		p, err := DecodePkg(bs)
		if err != nil {
			t.Error(err)
			return
		}

		t.Logf("%#v", p)

		buf := bufio.NewReader(bytes.NewReader(bs))

		t.Log(ReadWithPkg(buf))

	}
}
