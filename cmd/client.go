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

func RunClient(cfg *config.Config) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, 
		//syscall.SIGSTOP, //syscall.SIGQUIT
		)
	wg := sync.WaitGroup{}
	// TODO implement dial to server from client
	logger.Instance().Info("client started", zap.String("host", cfg.Host), zap.Int("port", cfg.Port))
	// TODO implement routines to send and to read messages
	<-done
	wg.Wait()
	logger.Instance().Info("client stopped")
}
