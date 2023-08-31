package io

import (
	"encoding/hex"
	"fmt"
	"log"
)

const (
	LevelWrite Level = 1 + iota
	LevelRead
	LevelInfo
	LevelDebug
	LevelError
)

type Level int

func (this Level) String() string {
	switch this {
	case LevelWrite:
		return "发送"
	case LevelRead:
		return "接收"
	case LevelInfo:
		return "信息"
	case LevelDebug:
		return "调试"
	case LevelError:
		return "错误"
	}
	return "未知"
}

func (this Level) printf(format string, v ...interface{}) {
	log.Printf("["+this.String()+"]"+format, fmt.Sprint(v...))
}

type ILog interface {
	Level(level Level)
	Readf(format string, v ...interface{})
	Writef(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

var Log = ILog(&_log{})

type _log struct {
	level  Level
	encode string
}

func (this *_log) WithASCII() {
	this.encode = "ascii"
}

func (this *_log) WithHEX() {
	this.encode = "hex"
}

func (this *_log) Level(level Level) {
	this.level = level
}

func (this *_log) Infof(format string, v ...interface{}) {
	LevelInfo.printf(format, v...)
}

func (this *_log) Writef(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if this.encode == "hex" {
		msg = hex.EncodeToString([]byte(msg))
	}
	LevelWrite.printf(msg)
}

func (this *_log) Readf(format string, v ...interface{}) {
	msg := fmt.Sprint(v...)
	if this.encode == "hex" {
		msg = hex.EncodeToString([]byte(msg))
	}
	LevelRead.printf(format, msg)
}

func (this *_log) Debugf(format string, v ...interface{}) {
	LevelDebug.printf(format, v...)
}

func (this *_log) Errorf(format string, v ...interface{}) {
	LevelError.printf(format, v...)
}
