package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"task4/internal/config"
)

var logger *zap.Logger

func Init(cfg *config.Config) {
	level := zapcore.InfoLevel 
	switch cfg.LogLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "dpanic":
		level = zapcore.DPanicLevel
	case "fatal":
		level = zapcore.FatalLevel
	case "panic":
		level = zapcore.PanicLevel
	}

	
	cfgZap := zap.NewProductionConfig()
	cfgZap.Level = zap.NewAtomicLevelAt(level)
	cfgZap.Encoding = "json" 

	
	l, err := cfgZap.Build()
	if err != nil {
		panic(err)
	}

	logger = l
}

func Instance() *zap.Logger {
	return logger
}
