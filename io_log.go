package io

import (
	"fmt"
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
	return &logger{
		Logger: l,
		level:  LevelAll,
		debug:  true,
		coding: "ascii",
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

func (this *logger) Debug(b ...bool) {
	this.debug = !(len(b) > 0 && !b[0])
}

func (this *logger) SetPrintWithHEX() {
	this.coding = "hex"
}

func (this *logger) SetPrintWithASCII() {
	this.coding = "ascii"
}

func (this *logger) Readln(prefix string, p []byte) {
	if this.debug && LevelRead >= this.level {
		switch this.coding {
		case "hex":
			this.Logger.Readf("%s%#X\n", prefix, p)
		case "ascii":
			this.Logger.Readf("%s%s\n", prefix, p)
		}
	}
}

func (this *logger) Writeln(prefix string, p []byte) {
	if this.debug && LevelWrite >= this.level {
		switch this.coding {
		case "hex":
			this.Logger.Writef("%s%#X\n", prefix, p)
		case "ascii":
			this.Logger.Writef("%s%s\n", prefix, p)
		}
	}
}

func (this *logger) Infof(format string, v ...interface{}) {
	if this.debug && LevelInfo >= this.level {
		this.Logger.Infof(format+"\n", v...)
	}
}

func (this *logger) Errorf(format string, v ...interface{}) {
	if this.debug && LevelError >= this.level {
		this.Logger.Errorf(format+"\n", v...)
	}
}

/*



 */

func NewLoggerWithWriter(w Writer) Logger {
	return &printer{w}
}

func NewLoggerWithStdout() Logger {
	return NewLoggerWithWriter(os.Stdout)
}

func NewLoggerWithChan(c chan []byte) Logger {
	return NewLoggerWithWriter(TryChan(c))
}

func NewLoggerChan() (Logger, chan []byte) {
	c := make(chan []byte, 10)
	return NewLoggerWithChan(c), c
}

type Logger interface {
	Readf(format string, v ...interface{})
	Writef(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

type printer struct{ Writer }

func (p printer) Readf(format string, v ...interface{})  { p.printf(LevelRead, format, v...) }
func (p printer) Writef(format string, v ...interface{}) { p.printf(LevelWrite, format, v...) }
func (p printer) Infof(format string, v ...interface{})  { p.printf(LevelInfo, format, v...) }
func (p printer) Errorf(format string, v ...interface{}) { p.printf(LevelError, format, v...) }

func (p printer) printf(level Level, format string, v ...interface{}) {
	timeStr := time.Now().Format("2006-01-02 15:04:05 ")
	p.Writer.Write([]byte(fmt.Sprintf(timeStr+"["+level.String()+"] "+format, v...)))
}
