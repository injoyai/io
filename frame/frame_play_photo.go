package frame

import (
	"errors"
	"github.com/injoyai/base/g"
	"github.com/injoyai/conv"
)

/*
UDP 最大可用字节为1472字节
以太网数据帧的长度必须在46-1500字节之间，这是由以太网的物理特性决定的，
这1500个字节被叫做链路层的MTU（最大传输单元)，IP首部为20个字节，
所以IP数据部分最大长度为1500-20=1480字节，
这1480个字节就是用来存放TCP或UDP数据包的,所以UDP数据报最大长度为1480，
UDP数据包的数据部分最大长度为1580-UDP首部8字节=1472字节

1张4K图片(32位)的大小理论为4096*2160*(32/8)= 35389440B = 33.75MB

所以理论上1张4K照片需要分成 35389440/1472=24041 个分片

实际2k大概 800~900KB ,约600个分片

Control
-----------------------------------------------------------------
|bit7	|bit6	|bit5	|bit4	|bit3	|bit2	|bit1	|bit0	|
-----------------------------------------------------------------
|																|
-----------------------------------------------------------------

*/

type PhotoStream struct {
	Control uint8  //控制字节,类型(基本信息,图片数据)
	No      uint8  //照片的序号取余
	Total   uint16 //总分片长度
	Index   uint16 //当前分片索引
	Data    []byte //数据内容
}

func (this *PhotoStream) Bytes() g.Bytes {
	data := []byte(nil)
	data = append(data, this.Control)
	data = append(data, this.No)
	data = append(data, conv.Bytes(this.Total)...)
	data = append(data, conv.Bytes(this.Index)...)
	data = append(data, this.Data...)
	return data
}

func (this *PhotoStream) IsStart() bool {
	return this.Index == 1 && this.Total > 0
}
func (this *PhotoStream) IsEnd() bool {
	return this.Index == this.Total && this.Total > 0
}

func DecodePhotoStream(bs []byte) (*PhotoStream, error) {
	if len(bs) < 6 {
		return nil, errors.New("数据长度不足")
	}
	return &PhotoStream{
		Control: bs[0],
		No:      bs[1],
		Total:   conv.Uint16(bs[2:4]),
		Index:   conv.Uint16(bs[4:6]),
		Data:    bs[6:],
	}, nil
}

/*



 */

type PhotoStreamCache struct {
	Info  map[string][]byte
	Photo map[uint16][]byte
}

func (this *PhotoStreamCache) Decode(bs []byte) error {
	p, err := DecodePhotoStream(bs)
	if err != nil {
		return err
	}
	switch p.Control {

	}
	return nil
}
