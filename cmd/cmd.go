package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"

	"task5/internal/config"
)

var ModeMap = map[string]func(config *config.Config) *Cmd{
	"":    NewApi,
	"api": NewApi,
}

type Cmd struct {
	authServer *fiber.App
	ctx        context.Context
	ctxDone    context.CancelFunc
}

func (c *Cmd) GetContext() context.Context {
	return c.ctx
}

// Wait Gracefully shutdown реализация
// программа будет ожитать сигнала от ОС на завершение и при поступлении его закроет все ресурсы
func (c *Cmd) Wait() {
	ch := make(chan os.Signal, 1)
	defer func() {
		ch <- os.Interrupt
	}()
	signal.Notify(ch,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	<-ch
	c.Close()
}

func (c *Cmd) Close() {
	c.ctxDone()
	err := c.authServer.Shutdown()
	if err != nil {
		panic("")
	}
}
