package io

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestNewDial(t *testing.T) {
	c := MustDial(func() (ReadWriteCloser, error) {
		return net.Dial("tcp", ":10089")
	})
	//if err != nil {
	//	t.Error(err)
	//	return
	//}

	c.Redial(func(ctx context.Context, c *Client) {
		c.Debug()
		c.SetPrintWithASCII()
		c.SetKey("test")
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					time.Sleep(time.Second)
					_, err := c.WriteString("666")
					if err != nil {
						return
					}
				}

			}
		}()
	})

	go func() {
		for {
			<-time.After(time.Second * 6)
			//c.Close()
			c.CloseAll()
		}
	}()

	go c.Run()
	go c.Run()

	select {}

}
