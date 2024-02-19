package bridge

import (
	"errors"
	"github.com/injoyai/base/maps/wait"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"log"
	"time"
)

type Client struct {
	*io.Client
	wait *wait.Entity
}

func (this *Client) Subscribe(Type string, port string) error {
	_, err := this.Write(io.NewSimple(io.SimpleControl{Type: io.OprSubscribe}, io.SimpleData{
		FliedListenType: []byte(Type),
		FliedListenPort: []byte(port),
	}, 3).Bytes())
	if err != nil {
		return err
	}
	_, err = this.wait.Wait("3")
	return err
}

func RedialClient(address string, option ...func(c *Client)) *Client {
	cli := &Client{
		wait: wait.New(time.Second),
	}
	dial.RedialTCP(address, func(c *io.Client) {
		cli.Client = c
		c.Logger.SetPrintWithUTF8()
		c.SetReadWriteWithSimple()
		c.SetKeepAlive(io.DefaultKeepAlive, io.NewSimplePing().Bytes())
		c.SetDealFunc(func(c *io.Client, msg io.Message) {
			p, err := io.DecodeSimple(msg)
			if err != nil {
				c.Logger.Errorf("decode bridge error: %v", err)
				return
			}

			switch p.Control.Type {
			case io.OprSubscribe:

				if p.Control.IsResponse {
					if p.Control.IsErr {
						cli.wait.Done("3", nil, errors.New(string(p.Data[io.FliedError])))
						return
					}
					cli.wait.Done("3", nil)
				}

			case io.OprWrite:

				log.Printf("[接收] [%s] %s", string(p.Data[io.FliedAddress]), string(p.Data[io.FliedData]))

			}
		})
		for _, v := range option {
			v(cli)
		}
	})
	return cli
}
