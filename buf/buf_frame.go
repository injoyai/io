package buf

import (
	"bufio"
	"encoding/hex"
	"errors"
	"github.com/injoyai/conv"
	"time"
)

type Frame struct {
	*StartEndFrame
	*LenFrame
	Timeout time.Duration //超时时间
}

func (this *Frame) ReadMessage(buf *bufio.Reader) ([]byte, error) {

	interval := time.Millisecond
	result := []byte(nil)

	for {
		b, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		result = append(result, b)

		//校验数据是否满足帧头帧尾
		seFull, err := this.StartEndFrame.Check(&result)
		if err != nil {
			return nil, err
		}

		//校验数据长度
		leFull, err := this.LenFrame.Check(result)
		if err != nil {
			return nil, err
		}

		//如果满足条件,则返回结果
		if seFull && leFull && !(this.StartEndFrame == nil && this.LenFrame == nil) {
			return result, nil
		}

		//未设置任何参数,读取全部数据
		if (this.StartEndFrame == nil || seFull) &&
			(this.LenFrame == nil || leFull) &&
			this.Timeout == 0 && buf.Buffered() == 0 {
			return result, nil
		}

		//根据超时时间结束读取
		waitTime := time.Duration(0)
		for buf.Buffered() == 0 && this.Timeout > 0 {
			<-time.After(interval)
			waitTime += interval
			if waitTime >= this.Timeout {
				return result, nil
			}
		}

	}

}

type StartEndFrame struct {
	Start, End []byte //帧头,帧尾
}

// Check 校验字节,响应数据完整性和错误
func (this *StartEndFrame) Check(bs *[]byte) (bool, error) {

	//未设置
	if this == nil {
		return true, nil
	}

	// 长度,减少代码长度...
	lenBs, lenStart, lenEnd := len(*bs), len(this.Start), len(this.End)

	//设置了帧头
	if lenStart > 0 {

		if lenBs <= lenStart {
			isStart := true
			for i, b := range *bs {
				if isStart {
					if b != this.Start[i] {
						isStart = false
					}
				}

				if !isStart {
					//寻找帧头
					if b == this.Start[0] {
						*bs = append((*bs)[:0], (*bs)[i:]...)
						break
					} else {
						*bs = append((*bs)[:0], (*bs)[:0]...)
						break
					}
				}

			}
		}

	}

	//基本数据长度不足
	if lenBs < lenStart+lenEnd {
		return false, nil
	}

	if lenEnd > 0 {

		//帧尾不符合,等待读取新的数据
		if hex.EncodeToString((*bs)[lenBs-lenEnd:]) != hex.EncodeToString(this.End) {
			return false, nil
		}

	}

	//未设置帧头,帧尾,任意数据皆满足条件
	return true, nil

}

type LenFrame struct {
	LittleEndian     bool //支持大端小端(默认false,大端),暂不支持2143,3412...
	LenStart, LenEnd uint //长度起始位置,长度结束位置
	LenFixed         int  //固定增加长度
}

func (this *LenFrame) Check(bs []byte) (bool, error) {

	//未设置
	if this == nil {
		return true, nil
	}

	//设置了错误的参数
	if this.LenStart > this.LenEnd {
		return false, errors.New("参数设置有误")
	}

	//数据还不满足条件
	if len(bs) <= int(this.LenEnd) {
		return false, nil
	}

	//获取数据总长度
	lenBytes := bs[this.LenStart : this.LenEnd+1]
	if this.LittleEndian {
		lenBytes = reverseBytes(lenBytes)
	}
	length := conv.Int(lenBytes) + this.LenFixed

	//数据异常,或设置的参数有误
	if length < len(bs) {
		return false, errors.New("数据长度过长")
	}

	//返回结果
	return length == len(bs), nil
}

// reverseBytes 字节数组倒序
func reverseBytes(bs []byte) []byte {
	x := make([]byte, len(bs))
	for i, v := range bs {
		x[len(bs)-i-1] = v
	}
	return x
}
