package main

import (
	"encoding/json"
	"fmt"
	"github.com/injoyai/io"
	"github.com/injoyai/io/extend/p2p"
	"github.com/injoyai/logs"
)

func main() {
	remoteAddr := "218.108.149.186:20000"
	_ = remoteAddr
	p, err := p2p.NewPeer(20001)
	if err != nil {
		logs.Err(err)
		return
	}
	p.SetDealFunc(func(msg *io.IMessage) {
		m := new(p2p.Msg)
		json.Unmarshal(msg.Bytes(), &m)
		switch m.Type {
		case p2p.TypeRegister:
			fmt.Println("register")
		}
	})

	logs.Err(p.Run())
}
