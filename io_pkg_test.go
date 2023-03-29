package io

import (
	"encoding/hex"
	"testing"
)

func TestNewPkg(t *testing.T) {

	data := []byte{0, 1, 2, 3, 5}
	{
		t.Log(NewPkg(20, data).Bytes().HEX()) //8888001280001400010203057909B04C8989

		s := "8888001280001400010203057909B04C8989"
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
		t.Log(NewPkg(20, data).SetCompress(1).Bytes().HEX()) //8888001280001400010203057909B04C8989

		s := "8888002A8800141F8B08000000000000FF626064626605040000FFFF5AE35D2605000000A491EE798989"
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

}
