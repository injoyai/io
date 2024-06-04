package dial

import (
	"bufio"
	"context"
	"github.com/injoyai/io"
	"time"
)

type Config struct {
	Dial           io.DialFunc   //连接函数
	Redial         bool          //重连
	RedialMaxNum   int           //重连未true有效,最大重连次数
	RedialMaxTime  time.Duration //重连未true有效,最大连接间隔,重连有效
	OnConnect      func(c *io.Client) error
	OnReadBuffer   func(buf *bufio.Reader) ([]byte, error)
	OnDealMessage  func(c *io.Client, msg io.Message)
	OnWriteMessage func(bs []byte) ([]byte, error)
	OnDisconnect   func(ctx context.Context, c *io.Client, err error)
	Options        []io.OptionClient
}

func WithConfig(cfg *Config) (*io.Client, error) {
	return WithConfigContext(context.Background(), cfg)
}

func WithConfigContext(ctx context.Context, cfg *Config) (*io.Client, error) {
	if cfg.Dial == nil {
		cfg.Dial = WithTCP(":10086")
	}
	fn := func(c *io.Client) {
		if cfg.OnConnect != nil {
			c.CloseWithErr(cfg.OnConnect(c))
		}
		if cfg.OnReadBuffer != nil {
			c.SetReadFunc(cfg.OnReadBuffer)
		}
		if cfg.OnDealMessage != nil {
			c.SetDealFunc(cfg.OnDealMessage)
		}
		if cfg.OnWriteMessage != nil {
			c.SetWriteFunc(cfg.OnWriteMessage)
		}
		if cfg.OnDisconnect != nil {
			c.SetCloseFunc(cfg.OnDisconnect)
		}
		if cfg.RedialMaxTime > 0 {
			c.SetRedialMaxTime(cfg.RedialMaxTime)
		}
		c.SetOptions(cfg.Options...)
	}
	if cfg.Redial {
		return io.RedialWithContext(ctx, cfg.Dial, fn), nil
	}
	return io.NewDialWithContext(ctx, cfg.Dial, fn)
}
