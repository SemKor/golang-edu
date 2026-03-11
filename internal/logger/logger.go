package logger

import (
	"go.uber.org/zap"

	"task4/internal/config"
)

var logger *zap.Logger

func Init(cfg *config.Config) {
	// TODO implement logger initialization func
}

func Instance() *zap.Logger {
	return logger
}
