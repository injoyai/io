package dial

import (
	"context"
	"github.com/injoyai/io"
)

func NewProxyServer(listen io.ListenFunc, dial io.DialFunc, options ...io.OptionServer) (*io.Server, error) {
	s, err := io.NewServer(listen)
	if err != nil {
		return nil, err
	}
	s.SetOptions(options...)
	s.SetReadWithKB(4)
	s.SetBeforeFunc(func(client *io.Client) error {
		_, err := io.NewDial(dial, func(c *io.Client) {
			c.Debug(false)
			c.SetReadWithKB(4)
			c.SetDealWithWriter(client)
			c.SetCloseFunc(func(ctx context.Context, msg *io.IMessage) {
				client.CloseWithErr(msg)
			})
			go c.Run()
			client.SetReadWithWriter(c)
		})
		return err
	})
	return s, nil
}

func RunProxyServer(listen io.ListenFunc, dial io.DialFunc, options ...io.OptionServer) error {
	s, err := NewProxyServer(listen, dial, options...)
	if err != nil {
		return err
	}
	return s.Run()
}
