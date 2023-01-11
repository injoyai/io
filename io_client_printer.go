package io

import "log"

func NewClientPrint() *ClientPrinter {
	cp := &ClientPrinter{}
	cp.SetPrintWithASCII()
	return cp
}

type ClientPrinter struct {
	debug     bool
	printFunc func(tag string, msg *Message)
}

// SetPrintFunc 设置打印函数
func (this *ClientPrinter) SetPrintFunc(fn func(tag string, msg *Message)) {
	this.printFunc = fn
}

// SetPrintWithHEX 设置打印HEX
func (this *ClientPrinter) SetPrintWithHEX() {
	this.printFunc = func(tag string, msg *Message) {
		log.Printf("[IO][%s] %s", tag, msg.HEX())
	}
}

// SetPrintWithASCII 设置打印ASCII
func (this *ClientPrinter) SetPrintWithASCII() {
	this.printFunc = func(tag string, msg *Message) {
		log.Printf("[IO][%s] %s", tag, msg.ASCII())
	}
}

// Print 打印输出
func (this *ClientPrinter) Print(tag string, msg *Message) {
	if this.debug && this.printFunc != nil {
		this.printFunc(tag, msg)
	}
}

// Debug 调试模式
func (this *ClientPrinter) Debug(b ...bool) {
	this.debug = !(len(b) > 0 && !b[0])
}
