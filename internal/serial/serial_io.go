package serial

import (
	"github.com/goburrow/serial"
	"io"
)

var _ io.ReadWriteCloser = (*Serial)(nil)

type Config = serial.Config

type Serial struct {
	serial.Port
}

func New(cfg *Config) (*Serial, error) {
	p, err := serial.Open(cfg)
	if err != nil {
		return nil, err
	}
	return &Serial{
		Port: p,
	}, nil
}
