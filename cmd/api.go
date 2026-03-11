package cmd

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"task5/internal/config"
	"task5/internal/logger"
	"task5/internal/transport/http"
)

// TODO добавить инициализацию слоя бизнес-логики и репозитория
func NewApi(cfg *config.Config) *Cmd {
	cmd := Cmd{}
	cmd.ctx = context.Background()
	cmd.ctx, cmd.ctxDone = context.WithCancel(cmd.ctx)
	logger.Init(&cfg.Log)
	lg := logger.Gist(cmd.ctx)
	server := http.NewServer(cfg)
	cmd.authServer = server.App
	err := server.App.Listen(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	if err != nil {
		lg.Fatal("cannot run http server", zap.Error(err))
	}
	return &cmd
}
