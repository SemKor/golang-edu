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

func RunServer(cfg *config.Config) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	var clients sync.Map

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Instance().Fatal("cannot start server", zap.Error(err))
	}
	defer ln.Close()

	logger.Instance().Info("server started",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
	)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				logger.Instance().Error("failed to accept connection", zap.Error(err))
				continue
			}

			clientID := conn.RemoteAddr().String()
			clients.Store(clientID, conn)

			wg.Add(1)
			go handleClient(clientID, conn, &clients, &wg)
		}
	}()

	<-done
	logger.Instance().Info("server stopping...")

	clients.Range(func(key, value interface{}) bool {
		conn := value.(net.Conn)
		conn.Close()
		return true
	})

	wg.Wait()

	logger.Instance().Info("server stopped")
}

func handleClient(clientID string, conn net.Conn, clients *sync.Map, wg *sync.WaitGroup) {
	defer func() {
		conn.Close()
		clients.Delete(clientID)
		logger.Instance().Info("client disconnected", zap.String("clientID", clientID))
		wg.Done()
	}()

	logger.Instance().Info("client connected", zap.String("clientID", clientID))

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()

		logger.Instance().Info("message received",
			zap.String("clientID", clientID),
			zap.String("message", msg),
		)

		broadcastMessage(clientID, msg, clients)
	}

	if err := scanner.Err(); err != nil {
		logger.Instance().Error("error reading",
			zap.String("clientID", clientID),
			zap.Error(err),
		)
	}
}

func broadcastMessage(senderID, message string, clients *sync.Map) {
	clients.Range(func(key, value interface{}) bool {
		clientID := key.(string)
		conn := value.(net.Conn)

		if clientID != senderID {
			_, err := fmt.Fprintf(conn, "%s: %s\n", senderID, message)
			if err != nil {
				logger.Instance().Error("failed to send message",
					zap.String("clientID", clientID),
					zap.Error(err),
				)
			}
		}
		return true
	})
}
