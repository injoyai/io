package io

import (
	"log"
)

func NewClientPrint() *ClientPrinter {
	cp := &ClientPrinter{prefix: "IO"}
	cp.SetPrintWithASCII()
	return cp
}

type ClientPrinter struct {
	prefix    string
	debug     bool
	printFunc func(prefix, tag, key string, msg Message)
}

// SetPrefix 设置前缀
func (this *ClientPrinter) SetPrefix(prefix string) {
	this.prefix = prefix
}

// SetPrintFunc 设置打印函数
func (this *ClientPrinter) SetPrintFunc(fn func(prefix, tag, key string, msg Message)) {
	this.printFunc = fn
}

// SetPrintWithHEX 设置打印HEX
func (this *ClientPrinter) SetPrintWithHEX() {
	this.printFunc = PrintWithHEX
}

// SetPrintWithASCII 设置打印ASCII
func (this *ClientPrinter) SetPrintWithASCII() {
	this.printFunc = PrintWithASCII
}

// Print 打印输出
func (this *ClientPrinter) Print(tag, key string, msg Message) {
	if this.debug && this.printFunc != nil {
		this.printFunc(this.prefix, tag, key, msg)
	}
}

// Debug 调试模式
func (this *ClientPrinter) Debug(b ...bool) {
	if this == nil {
		this = NewClientPrint()
	}
	this.debug = !(len(b) > 0 && !b[0])
}

func PrintWithHEX(prefix, tag, key string, msg Message) {
	switch tag {
	case TagRead, TagWrite:
		log.Printf("[%s][%s][%s] %s", prefix, tag, key, msg.HEX())
	default:
		log.Printf("[%s][%s][%s] %s", prefix, tag, key, msg.ASCII())
	}
}

func PrintWithASCII(prefix, tag, key string, msg Message) {
	log.Printf("[%s][%s][%s] %s", prefix, tag, key, msg.ASCII())
}
