package io

import (
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

type ILog interface {
	Level(level Level)
	Read(v ...interface{})
	Write(v ...interface{})
	Info(v ...interface{})
	Debug(v ...interface{})
	Error(v ...interface{})
}

var Log = ILog(&_log{})

type _log struct {
	level Level
}

func (this *_log) Level(level Level) {
	this.level = level
}

func (this *_log) Info(v ...interface{}) {
	this.print(LevelInfo, v...)
}

func (this *_log) Write(v ...interface{}) {
	this.print(LevelWrite, v...)
}

func (this *_log) Read(v ...interface{}) {
	this.print(LevelRead, v...)
}

func (this *_log) Debug(v ...interface{}) {
	this.print(LevelDebug, v...)
}

func (this *_log) Error(v ...interface{}) {
	this.print(LevelError, v...)
}

func (this *_log) print(level Level, v ...interface{}) {
	if this.level > 0 && this.level < level {
		return
	}
	log.Println(append([]interface{}{"[" + level.String() + "]"}, v...))
}
