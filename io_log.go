package io

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/injoyai/logs"
	"os"
	"time"
)

const (
	LevelAll Level = iota
	LevelWrite
	LevelRead
	LevelInfo
	LevelError
	LevelNone Level = 999
)

var (
	nullLogger = NewLoggerWithWriter(&null{})
)

type Level int

func (this Level) Name() string {
	switch this {
	case LevelWrite:
		return "发送"
	case LevelRead:
		return "接收"
	case LevelInfo:
		return "信息"
	case LevelError:
		return "错误"
	}
	return ""
}

func defaultLogger() *logger {
	return newLogger(NewLoggerWithStdout())
}

func NewLogger(l Logger) *logger {
	return newLogger(l)
}

func newLogger(l Logger) *logger {
	if l2, ok := l.(*logger); ok {
		return l2
	}
	if l == nil {
		l = nullLogger
	}
	return &logger{
		Logger: l,
		level:  LevelAll,
		debug:  true,
		coding: nil,
	}
}

type logger struct {
	Logger
	debug  bool                  //是否打印调试
	level  Level                 //日志等级
	coding func(p []byte) string //编码
}

func (this *logger) Debug(b ...bool) {
	this.debug = len(b) == 0 || b[0]
}

func (this *logger) SetLevel(level Level) {
	this.level = level
}

// SetPrintCoding 设置字节编码方式
func (this *logger) SetPrintCoding(coding func(p []byte) string) {
	this.coding = coding
}

// SetPrintWithHEX 设置字节编码方式hex
func (this *logger) SetPrintWithHEX() {
	this.SetPrintCoding(func(p []byte) string { return hex.EncodeToString(p) })
}

// SetPrintWithUTF8 设置字节编码方式utf-8
func (this *logger) SetPrintWithUTF8() {
	this.SetPrintCoding(func(p []byte) string { return string(p) })
}

func (this *logger) SetPrintWithBase64() {
	this.SetPrintCoding(func(p []byte) string { return base64.StdEncoding.EncodeToString(p) })
}

// Readln 打印读取到的数据
func (this *logger) Readln(prefix string, p []byte) {
	if this.debug && LevelRead >= this.level {
		if this.coding == nil {
			this.SetPrintWithUTF8()
		}
		this.Logger.Readf("%s%s\n", prefix, this.coding(p))
	}
}

// Writeln 打印写入的数据
func (this *logger) Writeln(prefix string, p []byte) {
	if this.debug && LevelWrite >= this.level {
		if this.coding == nil {
			this.SetPrintWithUTF8()
		}
		this.Logger.Writef("%s%s\n", prefix, this.coding(p))
	}
}

// Infof 打印信息
func (this *logger) Infof(format string, v ...interface{}) {
	if this.debug && LevelInfo >= this.level {
		this.Logger.Infof(format, v...)
	}
}

// Errorf 打印错误
func (this *logger) Errorf(format string, v ...interface{}) {
	if this.debug && LevelError >= this.level {
		this.Logger.Errorf(format, v...)
	}
}

// Printf 自定义打印
func (this *logger) Printf(format string, v ...interface{}) {
	if this.debug {
		this.Logger.Printf(format, v...)
	}
}

/*



 */

// NewLoggerWithWriter 新建输出到writer的日志
func NewLoggerWithWriter(w Writer) Logger {
	if w == nil {
		w = &null{}
	}
	return &logWriter{w}
}

// NewLoggerWithStdout 新建输出到终端的日志
func NewLoggerWithStdout() Logger {
	return NewLoggerWithWriter(os.Stdout)
}

// NewLoggerWithChan 新建输出到指定通道的日志
func NewLoggerWithChan(c chan []byte) Logger {
	return NewLoggerWithWriter(TryChan(c))
}

// NewLoggerChan 新建输出到通道的日志
func NewLoggerChan() (Logger, chan []byte) {
	c := make(chan []byte, 10)
	return NewLoggerWithChan(c), c
}

// NullLogger 无输出的日志
func NullLogger() Logger {
	return nullLogger
}

type Logger interface {
	Readf(format string, v ...interface{})
	Writef(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Printf(format string, v ...interface{})
}

//==========================================logWriter==========================================

type logWriter struct{ Writer }

func (p logWriter) Readf(format string, v ...interface{})  { p.printf(LevelRead, format, v...) }
func (p logWriter) Writef(format string, v ...interface{}) { p.printf(LevelWrite, format, v...) }
func (p logWriter) Infof(format string, v ...interface{})  { p.printf(LevelInfo, format, v...) }
func (p logWriter) Errorf(format string, v ...interface{}) { p.printf(LevelError, format, v...) }

func (p logWriter) Printf(format string, v ...interface{}) {
	if p.Writer == nil {
		return
	}
	_, err := p.Writer.Write([]byte(fmt.Sprintf(format, v...)))
	logs.PrintErr(err)
}

func (p logWriter) printf(level Level, format string, v ...interface{}) {
	if p.Writer == nil {
		return
	}
	timeStr := time.Now().Format("2006-01-02 15:04:05 ")
	_, err := p.Writer.Write([]byte(fmt.Sprintf(timeStr+"["+level.Name()+"] "+format, v...)))
	logs.PrintErr(err)
}
