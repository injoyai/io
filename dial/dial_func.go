package dial

import (
	"github.com/injoyai/io"
	"github.com/injoyai/logs"
)

func init() {
	//设置只打印到终端
	logs.SetWriter(logs.Stdout)
}

func RunServer(s *io.Server, err error) error {
	if err != nil {
		return err
	}
	return s.Run()
}
