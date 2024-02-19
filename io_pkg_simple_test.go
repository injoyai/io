package io

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"github.com/injoyai/logs"
	"testing"
)

func TestDecodeSimple(t *testing.T) {
	s := "6800370200096c697374656e4b6579097463702e3130303836026970093132372e302e302e3104706f7274053530353138036d7367033333332f"
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
			Type: OprRead,
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

func TestReadWithSimple(t *testing.T) {
	s := "68002403030a6c697374656e54797065037463700a6c697374656e506f727405313030383659"
	bs, _ := hex.DecodeString(s)
	bs, err := ReadWithSimple(bufio.NewReader(bytes.NewReader(bs)))
	if err != nil {
		t.Error(err)
		//return
	}
	logs.Debug(hex.EncodeToString(bs))
}
