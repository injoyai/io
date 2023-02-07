package pipe

import (
	"github.com/injoyai/io"
	"log"
)

func newDealFunc(dealFunc func(msg *Message) error) func(msg *io.ClientMessage) {
	return func(msg *io.ClientMessage) {
		result, err := decodeMessage(msg.Bytes())
		if err != nil {
			log.Println("[错误]", err)
			return
		}

		if dealFunc != nil {
			if err := dealFunc(result); err != nil {
				log.Println("[错误]", err)
				return
			}
		}
	}
}

func writeFunc(writeLen uint, writeBytes func(p []byte) (int, error), key, addr string, p []byte) (int, error) {

	total := len(p)

	//一次性发送全部数据,数据太多可能会影响到其他数据的时效
	if writeLen == 0 {
		msg := newWriteMessage(key, addr, p)
		return writeBytes(msg.Bytes())
	}

	//分包发送,避免其他数据不能及时发送
	for len(p) > 0 {
		data := []byte(nil)
		if len(p) > int(writeLen) {
			data = p[:writeLen]
			p = p[writeLen:]
		} else {
			data = p[:]
			p = p[:0]
		}
		msg := newWriteMessage(key, addr, data)
		if _, err := writeBytes(msg.Bytes()); err != nil {
			return 0, err
		}
	}

	return total, nil
}
