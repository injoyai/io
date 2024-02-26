package listen

import (
	"github.com/injoyai/io"
)

// NewProxyServer 监听连接,并代理新连接
func NewProxyServer(listen io.ListenFunc, dial io.DialFunc, options ...io.OptionServer) (*io.Server, error) {
	return io.NewServer(listen, func(s *io.Server) {
		s.Logger.Debug(false)
		s.SetOptions(options...)
		s.SetReadWith1KB()
		s.SetBeforeFunc(func(client *io.Client) error {
			_, err := io.NewDial(dial, func(c *io.Client) {
				c.Debug(false)
				io.SwapClient(c, client)
			})
			return err
		})
	})
}

// RunProxyServer 监听连接,并代理新连接
func RunProxyServer(listen io.ListenFunc, dial io.DialFunc, options ...io.OptionServer) error {
	s, err := NewProxyServer(listen, dial, options...)
	if err != nil {
		return err
	}
	return s.Run()
}
