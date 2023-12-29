package buf

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"testing"
)

func TestNewReadWithStartEnd(t *testing.T) {
	{
		fn := NewReadWithStartEnd([]byte{0x03, 0x03}, []byte{0x04, 0x04})
		buf := bufio.NewReader(bytes.NewBuffer([]byte{0x03, 0x03, 0x11, 0x011, 0x04, 0x04, 0x05}))
		val, err := fn(buf)
		if err != nil {
			t.Error(err)
		}
		if hex.EncodeToString(val) != hex.EncodeToString([]byte{0x03, 0x03, 0x11, 0x011, 0x04, 0x04}) {
			t.Error("测试失败" + hex.EncodeToString(val))
		} else {
			t.Log("测试通过")
		}

	}
	{
		fn := NewReadWithStartEnd([]byte{0x03, 0x03}, []byte{0x04, 0x04})
		buf := bufio.NewReader(bytes.NewBuffer([]byte{0x01, 0x03, 0x03, 0x11, 0x011, 0x04, 0x04, 0x05}))
		val, err := fn(buf)
		if err != nil {
			t.Error(err)
		}
		if hex.EncodeToString(val) != hex.EncodeToString([]byte{0x03, 0x03, 0x11, 0x011, 0x04, 0x04}) {
			t.Error("测试失败" + hex.EncodeToString(val))
		} else {
			t.Log("测试通过")
		}
	}
	{
		fn := NewReadWithLen(&LenFrame{
			LenStart: 1,
			LenEnd:   1,
			LenFixed: 3,
		})
		buf := bufio.NewReader(bytes.NewBuffer([]byte{0x01, 0x03, 0x03, 0x11, 0x011, 0x04, 0x04, 0x05}))
		val, err := fn(buf)
		if err != nil {
			t.Error(err)
		}
		if hex.EncodeToString(val) != hex.EncodeToString([]byte{0x01, 0x03, 0x03, 0x11, 0x011, 0x04}) {
			t.Error("测试失败" + hex.EncodeToString(val))
		} else {
			t.Log("测试通过")
		}
	}
	{
		fn := &Frame{
			StartEndFrame: &StartEndFrame{
				Start: []byte{0x03},
				End:   nil,
			},
			LenFrame: &LenFrame{
				LenStart: 1,
				LenEnd:   1,
				LenFixed: 3,
			},
			Timeout: 0,
		}
		buf := bufio.NewReader(bytes.NewBuffer([]byte{0x01, 0x03, 0x03, 0x11, 0x011, 0x04, 0x04, 0x05}))
		val, err := fn.ReadMessage(buf)
		if err != nil {
			t.Error(err)
		}
		if hex.EncodeToString(val) != hex.EncodeToString([]byte{0x03, 0x03, 0x11, 0x011, 0x04, 0x04}) {
			t.Error("测试失败" + hex.EncodeToString(val))
		} else {
			t.Log("测试通过")
		}
	}
}

func TestReadPrefix(t *testing.T) {
	buf := bufio.NewReader(bytes.NewBuffer([]byte{0x03, 0x11, 0x11, 0x03, 0x03, 0x11, 0x011, 0x04, 0x04, 0x05}))
	bs, err := ReadPrefix(buf, []byte{0x03, 0x03, 0x11})
	if err != nil {
		t.Error(err)
	}
	t.Log(hex.EncodeToString(bs))
	bs, err = ReadWithAll(buf)
	if err != nil {
		t.Error(err)
	}
	t.Log(hex.EncodeToString(bs))
}
