package service

import "go.uber.org/zap"

type Lvl uint8

const (
	DEBUG Lvl = iota + 1
	INFO
	WARN
	ERROR
)

type Logger interface {
	Debug(string, ...zap.Field)
	Info(string, ...zap.Field)
	Warn(string, ...zap.Field)
	Error(string, ...zap.Field)
	Fatal(string, ...zap.Field)
	Panic(string, ...zap.Field)
	SetLevel(Lvl)
	Zap() *zap.Logger
}
