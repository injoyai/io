package io

import (
	"encoding/hex"
	golog "log"
)

const (
	LevelAll Level = iota
	LevelWrite
	LevelRead
	LevelDebug
	LevelInfo
	LevelError
	LevelNone Level = 999
)

type Level int

var (
	LevelMap = map[Level]string{
		LevelWrite: "发送",
		LevelRead:  "接收",
		LevelDebug: "调试",
		LevelInfo:  "信息",
		LevelError: "错误",
	}
)

func (this Level) String() string {
	return LevelMap[this]
}

func (this Level) printf(log *log, format string, v ...interface{}) {
	if log.debug && this >= log.level {
		golog.Printf("["+this.String()+"]"+format, v...)
	}
}

type Logger interface {
	SetLevel(level Level)
	SetPrintWithHEX()
	SetPrintWithASCII()
	Debug(b ...bool)
	Readln(prefix string, p []byte)
	Writeln(prefix string, p []byte)
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

func NewLog() Logger {
	return &log{
		level:  LevelAll,
		debug:  true,
		coding: "ascii",
	}
}

type log struct {
	level  Level
	debug  bool
	coding string
}

func (this *log) SetLevel(level Level) {
	this.level = level
}

func (this *log) Debug(b ...bool) {
	this.debug = !(len(b) > 0 && !b[0])
}

func (this *log) SetPrintWithHEX() {
	this.coding = "hex"
}

func (this *log) SetPrintWithASCII() {
	this.coding = "ascii"
}

func (this *log) Writeln(prefix string, p []byte) {
	switch this.coding {
	case "hex":
		LevelWrite.printf(this, "%s%s", prefix, hex.EncodeToString(p))
	default:
		LevelWrite.printf(this, "%s%s", prefix, string(p))
	}
}

func (this *log) Readln(prefix string, p []byte) {
	switch this.coding {
	case "hex":
		LevelRead.printf(this, "%s%s", prefix, hex.EncodeToString(p))
	default:
		LevelRead.printf(this, "%s%s", prefix, string(p))
	}
}

func (this *log) Infof(format string, v ...interface{}) {
	LevelInfo.printf(this, format, v...)
}

func (this *log) Debugf(format string, v ...interface{}) {
	LevelDebug.printf(this, format, v...)
}

func (this *log) Errorf(format string, v ...interface{}) {
	LevelError.printf(this, format, v...)
}
