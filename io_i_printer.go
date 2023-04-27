package io

import (
	"fmt"
	"io"
	"log"
	"strings"
)

func newPrinter(key string) *printer {
	cp := &printer{}
	cp.SetKey(key)
	cp.SetPrintWithASCII()
	return cp
}

type printer struct {
	key       string
	debug     bool
	printFunc PrintFunc
}

// SetKey 设置唯一标识
func (this *printer) SetKey(key string) {
	this.key = key
}

// GetKey 获取唯一标识
func (this *printer) GetKey() string {
	return this.key
}

// SetPrintFunc 设置打印函数
func (this *printer) SetPrintFunc(fn PrintFunc) {
	this.printFunc = fn
}

// GetPrintFunc 获取打印函数
func (this *printer) GetPrintFunc() PrintFunc {
	return this.printFunc
}

// SetPrintWithHEX 设置打印HEX
func (this *printer) SetPrintWithHEX() {
	this.printFunc = PrintWithHEX
}

// SetPrintWithASCII 设置打印ASCII
func (this *printer) SetPrintWithASCII() {
	this.printFunc = PrintWithASCII
}

// Print 打印输出
func (this *printer) Print(msg Message, tag ...string) {
	if this.debug && this.printFunc != nil {
		this.printFunc(msg, tag...)
	}
}

// Debug 调试模式
// 开启时,已连接的客户端无效,需重连生效
func (this *printer) Debug(b ...bool) {
	this.debug = !(len(b) > 0 && !b[0])
}

// GetDebug 获取调试
func (this *printer) GetDebug() bool {
	return this.debug
}

//================================Common================================

func PrintWithHEX(msg Message, tag ...string) {
	log.Print(PrintfWithHEX(msg, tag...))
}

func PrintWithASCII(msg Message, tag ...string) {
	log.Print(PrintfWithASCII(msg, tag...))
}

func PrintWithBase(msg Message, tag ...string) {
	if len(tag) > 0 && (tag[0] == TagErr || tag[0] == TagInfo) {
		log.Print(PrintfWithASCII(msg, tag...))
	}
}

func PrintWithErr(msg Message, tag ...string) {
	if len(tag) > 0 && tag[0] == TagErr && msg.String() != io.EOF.Error() {
		log.Print(PrintfWithASCII(msg, tag...))
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
	return fmt.Sprintf("%s %s\n", t, msg.ASCII())
}

func PrintfWithASCII(msg Message, tag ...string) string {
	t := strings.Join(tag, "][")
	if len(t) > 0 {
		t = "[" + t + "]"
	}
	return fmt.Sprintf("%s %s\n", t, msg.ASCII())
}
