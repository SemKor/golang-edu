package cmd

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"task5/internal/config"
	"task5/internal/domain"
	"task5/internal/logger"
	"task5/internal/repository/postgres"
	"task5/internal/transport/http"
)

// TODO расширить postgres repository остальными методами
func NewApi(cfg *config.Config) *Cmd {
	cmd := Cmd{}
	cmd.ctx = context.Background()
	cmd.ctx, cmd.ctxDone = context.WithCancel(cmd.ctx)

	logger.Init(&cfg.Log)
	lg := logger.Gist(cmd.ctx)

	db, err := postgres.NewDB(cfg)
	if err != nil {
		lg.Fatal("cannot connect to database", zap.Error(err))
	}

	repo := postgres.NewRepository(db)

	usecase := domain.NewUsecase(
		repo,
		repo,
		repo,
		repo,
		repo,
	)

	server := http.NewServer(cfg, usecase)
	cmd.authServer = server.App

	err = server.App.Listen(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	if err != nil {
		lg.Fatal("cannot run http server", zap.Error(err))
	}

	return &cmd
}