package io

import (
	"fmt"
	"log"
	"strings"
)

func NewIPrinter() *IPrinter {
	cp := &IPrinter{}
	cp.SetPrintWithASCII()
	return cp
}

type IPrinter struct {
	debug     bool
	printFunc PrintFunc
}

// SetPrintFunc 设置打印函数
func (this *IPrinter) SetPrintFunc(fn PrintFunc) {
	this.printFunc = fn
}

// SetPrintWithHEX 设置打印HEX
func (this *IPrinter) SetPrintWithHEX() {
	this.printFunc = PrintWithHEX
}

// SetPrintWithASCII 设置打印ASCII
func (this *IPrinter) SetPrintWithASCII() {
	this.printFunc = PrintWithASCII
}

// Print 打印输出
func (this *IPrinter) Print(msg Message, tag ...string) {
	if this.debug && this.printFunc != nil {
		this.printFunc(msg, tag...)
	}
}

// Debug 调试模式
func (this *IPrinter) Debug(b ...bool) {
	this.debug = !(len(b) > 0 && !b[0])
}

//================================Common================================

func PrintWithHEX(msg Message, tag ...string) {
	log.Print(PrintfWithHEX(msg, tag...))
}

func PrintWithASCII(msg Message, tag ...string) {
	log.Print(PrintfWithASCII(msg, tag...))
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

func PrintfWithASCII(msg Message, tag ...string) string {
	t := strings.Join(tag, "][")
	if len(t) > 0 {
		t = "[" + t + "]"
	}
	return fmt.Sprintf("%s %s", t, msg.ASCII())
}
