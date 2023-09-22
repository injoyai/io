package listen

import (
	"github.com/injoyai/io"
)

func RunServer(s *io.Server, err error, options ...io.OptionServer) error {
	if err != nil {
		return err
	}
	return s.SetOptions(options...).Run()
}
