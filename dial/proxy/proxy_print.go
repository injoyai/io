package proxy

import (
	"github.com/injoyai/io"
	"github.com/injoyai/logs"
)

func PrintWithASCII(msg io.Message, tag ...string) {
	switch true {
	case msg.String() == io.Ping || msg.String() == io.Pong:
	case len(tag) > 0:
		str := msg.ASCII()
		switch tag[0] {
		case io.TagWrite, io.TagRead:
			m, err := DecodeMessage(msg.Bytes())
			if err != nil {
				logs.Debug(err, ":", msg.ASCII())
				return
			}
			str = m.Bytes().ASCII()
		}
		logs.Debugf(io.PrintfWithASCII([]byte(str), append([]string{"PI|C"}, tag...)...))
	}
}
