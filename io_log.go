package io

import (
	"encoding/hex"
	golog "log"
)

const (
	LevelWrite Level = 1 + iota
	LevelRead
	LevelDebug
	LevelInfo
	LevelError
)

type Level int

func (this Level) String() string {
	switch this {
	case LevelWrite:
		return "发送"
	case LevelRead:
		return "接收"
	case LevelDebug:
		return "调试"
	case LevelInfo:
		return "信息"
	case LevelError:
		return "错误"
	}
	return "未知"
}

func (this Level) printf(log *log, format string, v ...interface{}) {
	if log.debug && this >= log.level {
		prefix := "[" + this.String() + "]"
		if len(log.key) > 0 {
			prefix += "[" + log.key + "]"
		}
		golog.Printf(prefix+format, v...)
	}
}

type Logger = ILog

type ILog interface {
	Level(level Level)
	Debug(b ...bool)
	Read(p []byte)
	Write(p []byte)
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

func newLog() Logger {
	return &log{
		debug:  true,
		coding: "ascii",
	}
}

type log struct {
	level  Level
	key    string
	debug  bool
	coding string
}

func (this *log) SetKey(key string) {
	this.key = key
}

func (this *log) Level(level Level) {
	this.level = level
}

func (this *log) Debug(b ...bool) {
	this.debug = !(len(b) > 0 && b[0])
}

func (this *log) SetPrintWithHEX() {
	this.coding = "hex"
}

func (this *log) SetPrintWithASCII() {
	this.coding = "ascii"
}

func (this *log) Write(p []byte) {
	switch this.coding {
	case "hex":
		LevelWrite.printf(this, "%s", hex.EncodeToString(p))
	default:
		LevelWrite.printf(this, "%s", string(p))
	}
}

func (this *log) Read(p []byte) {
	switch this.coding {
	case "hex":
		LevelRead.printf(this, "%s", hex.EncodeToString(p))
	default:
		LevelRead.printf(this, "%s", string(p))
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
