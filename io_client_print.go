package io

import "log"

type ClientPrint struct {
	printFunc func(tag string, msg *Message)
}

// SetPrintFunc 设置打印函数
func (this *ClientPrint) SetPrintFunc(fn func(tag string, msg *Message)) {
	this.printFunc = fn
}

// SetPrintWithHEX 设置打印HEX
func (this *ClientPrint) SetPrintWithHEX() {
	this.printFunc = func(tag string, msg *Message) {
		log.Printf("[IO][%s] %s", tag, msg.HEX())
	}
}

// SetPrintWithASCII 设置打印ASCII
func (this *ClientPrint) SetPrintWithASCII() {
	this.printFunc = func(tag string, msg *Message) {
		log.Printf("[IO][%s] %s", tag, msg.ASCII())
	}
}
