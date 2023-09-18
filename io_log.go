package io

import (
	"encoding/hex"
	"github.com/injoyai/conv"
	"github.com/injoyai/logs"
)

const (
	LevelAll Level = iota
	LevelWrite
	LevelRead
	LevelDebug
	LevelInfo
	LevelError
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
		prefix := "[" + this.String() + "]"
		if len(log.key) > 0 {
			prefix += "[" + log.key + "]"
		}
		logs.New(this.String()).Writef(logs.LevelError, prefix+" "+format, v...)
	}
}

type Logger interface {
	SetLevel(level Level)
	GetKey() string
	SetKey(key string)
	SetPrintWithHEX()
	SetPrintWithASCII()
	Debug(b ...bool)
	Readln(p []byte)
	Writeln(p []byte)
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

var NewLog = newLog

func newLog(key ...string) Logger {
	for _, name := range LevelMap {
		logs.New(name).SetName("")
	}
	return &log{
		level:  LevelAll,
		key:    conv.DefaultString("", key...),
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

func (this *log) GetKey() string {
	return this.key
}

func (this *log) SetKey(key string) {
	this.key = key
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

func (this *log) Writeln(p []byte) {
	switch this.coding {
	case "hex":
		LevelWrite.printf(this, "%s", hex.EncodeToString(p))
	default:
		LevelWrite.printf(this, "%s", string(p))
	}
}

func (this *log) Readln(p []byte) {
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
