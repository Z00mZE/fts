package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Z00mZE/fts/domain/service"
)

type ZapLogger struct {
	*zap.Logger
	lvl zap.AtomicLevel
}

func (z *ZapLogger) Zap() *zap.Logger {
	return z.Logger
}

func NewZapLogger() *ZapLogger {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevel()
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.ErrorOutputPaths = []string{"stderr"}
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	zapConfig.EncoderConfig.CallerKey = ""

	log, _ := zapConfig.Build(
		zap.AddCaller(),
	)
	return &ZapLogger{log, zapConfig.Level}
}

func (z *ZapLogger) SetLevel(lvl service.Lvl) {
	switch lvl {
	case service.DEBUG:
		z.lvl.SetLevel(zapcore.DebugLevel)
	case service.INFO:
		z.lvl.SetLevel(zapcore.InfoLevel)
	case service.WARN:
		z.lvl.SetLevel(zapcore.WarnLevel)
	case service.ERROR:
		z.lvl.SetLevel(zapcore.ErrorLevel)
	}
}
