package main

import (
	"bytes"
	"github.com/injoyai/io"
	"github.com/injoyai/logs"
)

func main() {
	buf1 := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	buf2 := bytes.NewReader([]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19})
	buf := io.MultiReader(buf1, buf2)
	bs, err := io.ReadAll(buf)
	logs.PrintErr(err)
	logs.Debug(bs)
}
