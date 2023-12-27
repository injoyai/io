package io

import (
	"encoding/base64"
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

const (
	codingHEX    = "hex"
	codingBase64 = "base64"
	codingUTF8   = "utf8"
)

type Level int

var (
	LevelMap = map[Level]string{
		LevelWrite: "发送",
		LevelRead:  "接收",
		LevelInfo:  "信息",
		LevelError: "错误",
	}
)

func (this Level) String() string {
	return LevelMap[this]
}

func defaultLogger() *logger {
	return newLogger(NewLoggerWithStdout())
}

func NewLogger(l Logger) *logger {
	return newLogger(l)
}

func newLogger(l Logger) *logger {
	if l == nil {
		l = NewLoggerNull()
	}
	if l2, ok := l.(*logger); ok {
		return l2
	}
	return &logger{
		Logger: l,
		level:  LevelAll,
		debug:  true,
		coding: codingUTF8,
	}
}

type logger struct {
	Logger
	level  Level  //日志等级
	debug  bool   //是否打印调试
	coding string //编码
}

func (this *logger) SetLevel(level Level) {
	this.level = level
}

// Debug 实现Debugger接口,不用返回值
func (this *logger) Debug(b ...bool) {
	this.debug = !(len(b) > 0 && !b[0])
}

// SetPrintWithHEX 设置字节编码方式hex
func (this *logger) SetPrintWithHEX() {
	this.coding = codingHEX
}

// SetPrintWithUTF8 设置字节编码方式utf-8
func (this *logger) SetPrintWithUTF8() {
	this.coding = codingUTF8
}

func (this *logger) SetPrintWithBase64() {
	this.coding = codingBase64
}

// Readln 打印读取到的数据
func (this *logger) Readln(prefix string, p []byte) {
	if this.debug && LevelRead >= this.level {
		switch this.coding {
		case codingHEX:
			this.Logger.Readf("%s%#x\n", prefix, p)
		case codingUTF8:
			this.Logger.Readf("%s%s\n", prefix, p)
		case codingBase64:
			this.Logger.Readf("%s%s\n", prefix, base64.StdEncoding.EncodeToString(p))
		}
	}
}

// Writeln 打印写入的数据
func (this *logger) Writeln(prefix string, p []byte) {
	if this.debug && LevelWrite >= this.level {
		switch this.coding {
		case codingHEX:
			this.Logger.Writef("%s%#x\n", prefix, p)
		case codingUTF8:
			this.Logger.Writef("%s%s\n", prefix, p)
		case codingBase64:
			this.Logger.Writef("%s%s\n", prefix, base64.StdEncoding.EncodeToString(p))
		}
	}
}

// Infof 打印信息
func (this *logger) Infof(format string, v ...interface{}) {
	if this.debug && LevelInfo >= this.level {
		this.Logger.Infof(format+"\n", v...)
	}
}

// Errorf 打印错误
func (this *logger) Errorf(format string, v ...interface{}) {
	if this.debug && LevelError >= this.level {
		this.Logger.Errorf(format+"\n", v...)
	}
}

/*



 */

// NewLoggerWithWriter 新建输出到writer的日志
func NewLoggerWithWriter(w Writer) Logger {
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

// NewLoggerNull 新建无输出的日志
func NewLoggerNull() Logger {
	return NewLoggerWithWriter(&null{})
}

type Logger interface {
	Readf(format string, v ...interface{})
	Writef(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

//==========================================logWriter==========================================

type logWriter struct{ Writer }

func (p logWriter) Readf(format string, v ...interface{})  { p.printf(LevelRead, format, v...) }
func (p logWriter) Writef(format string, v ...interface{}) { p.printf(LevelWrite, format, v...) }
func (p logWriter) Infof(format string, v ...interface{})  { p.printf(LevelInfo, format, v...) }
func (p logWriter) Errorf(format string, v ...interface{}) { p.printf(LevelError, format, v...) }

func (p logWriter) printf(level Level, format string, v ...interface{}) {
	timeStr := time.Now().Format("2006-01-02 15:04:05 ")
	_, err := p.Writer.Write([]byte(fmt.Sprintf(timeStr+"["+level.String()+"] "+format, v...)))
	logs.PrintErr(err)
}
