package io

import (
	"fmt"
	"log"
	"strings"
)

func NewClientPrint() *ClientPrinter {
	cp := &ClientPrinter{}
	cp.SetPrintWithASCII()
	return cp
}

type ClientPrinter struct {
	debug     bool
	printFunc PrintFunc
}

// SetPrintFunc 设置打印函数
func (this *ClientPrinter) SetPrintFunc(fn PrintFunc) {
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
func (this *ClientPrinter) Print(msg Message, tag ...string) {
	if this.debug && this.printFunc != nil {
		this.printFunc(msg, tag...)
	}
}

// Debug 调试模式
func (this *ClientPrinter) Debug(b ...bool) {
	if this == nil {
		this = NewClientPrint()
	}
	this.debug = !(len(b) > 0 && !b[0])
}

func PrintWithHEX(msg Message, tag ...string) {
	t := strings.Join(tag, "][")
	if len(t) > 0 {
		t = "[" + t + "]"
	}
	if strings.Contains(t, TagRead) || strings.Contains(t, TagWrite) {
		log.Printf("%s %s", t, msg.HEX())
	} else {
		log.Printf("%s %s", t, msg.ASCII())
	}
}

func PrintfWithHEX(msg Message, tag ...string) string {
	t := strings.Join(tag, "][")
	if len(t) > 0 {
		t = "[" + t + "]"
	}
	if strings.Contains(t, TagRead) || strings.Contains(t, TagWrite) {
		return fmt.Sprintf("%s %s", t, msg.HEX())
	}
	return fmt.Sprintf("%s %s", t, msg.ASCII())
}

func PrintWithASCII(msg Message, tag ...string) {
	t := strings.Join(tag, "][")
	if len(t) > 0 {
		t = "[" + t + "]"
	}
	log.Printf("%s %s", t, msg.ASCII())
}

func PrintfWithASCII(msg Message, tag ...string) string {
	t := strings.Join(tag, "][")
	if len(t) > 0 {
		t = "[" + t + "]"
	}
	return fmt.Sprintf("%s %s", t, msg.ASCII())
}
