package cmd

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"

	"task4/internal/config"
	"task4/internal/logger"
)

func RunServer(cfg *config.Config) {
	done := make(chan os.Signal, 1)
	wg := sync.WaitGroup{}
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, 
		//syscall.SIGSTOP, syscall.SIGQUIT
	)
	// TODO implement start server to listen to port here
	logger.Instance().Info("server started", zap.String("host", cfg.Host), zap.Int("port", cfg.Port))
	// TODO implement routines to process connections here
	<-done
	wg.Wait()
	logger.Instance().Info("server stopped")
}
