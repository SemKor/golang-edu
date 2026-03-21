package cmd

import (
	"bufio"
	"fmt"
	"net"
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
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := sync.WaitGroup{}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		logger.Instance().Fatal("cannot connect to server", zap.Error(err))
		return
	}
	defer conn.Close()

	logger.Instance().Info("client started", zap.String("host", cfg.Host), zap.Int("port", cfg.Port), zap.String("clientName", cfg.ClientName))

	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			msg := scanner.Text()
			fmt.Println(msg) 
		}
		if err := scanner.Err(); err != nil {
			logger.Instance().Error("error reading from server", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		input := bufio.NewScanner(os.Stdin)
		for input.Scan() {
			text := input.Text()
			if text == "" {
				continue
			}
			msg := fmt.Sprintf("%s: %s", cfg.ClientName, text)
			_, err := fmt.Fprintln(conn, msg)
			if err != nil {
				logger.Instance().Error("failed to send message", zap.Error(err))
			}
		}
		if err := input.Err(); err != nil {
			logger.Instance().Error("error reading from stdin", zap.Error(err))
		}
	}()

	<-done
	logger.Instance().Info("client stopping")
	conn.Close()
	wg.Wait()
}