package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"task5/internal/config"
)

type ctxLogger struct{}

var gist *zap.Logger

func Gist(ctx context.Context) *zap.Logger {
	l, ok := ctx.Value(ctxLogger{}).(*zap.Logger)
	if ok {
		return l
	}
	return gist
}

func WithLogger(ctx context.Context, lg *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxLogger{}, lg)
}

func SetLogger(lg *zap.Logger, f func(key any, value any)) {
	f(ctxLogger{}, lg)
}

func Init(cfg *config.Log) {
	lvlerr := false
	fmterr := false
	cfgLvl, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		cfgLvl = zapcore.DebugLevel
		lvlerr = true
	}
	lvlEnablerFn := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= cfgLvl
	})

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	var encoder zapcore.Encoder
	switch cfg.Format {
	case "text":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	case "json", "":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
		fmterr = true
	}
	output := zapcore.Lock(os.Stdout)
	gist = zap.New(zapcore.NewCore(encoder, output, lvlEnablerFn))
	gist.Named(cfg.Title)
	_ = zap.ReplaceGlobals(gist)
	if lvlerr {
		gist.Error("log level specified in config is invalid, error is assumed")
	}
	if fmterr {
		gist.Error("formatter specified in config is invalid, json is assumed")
	}
}
