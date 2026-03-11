package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"task5/internal/config"
)

type Server struct {
	App *fiber.App
}

// NewServer TODO закончить инициализацию сервера обработчиками
func NewServer(cfg *config.Config) *Server {
	instance := Server{
		App: fiber.New(),
	}
	h := newHandler(cfg)
	instance.App.Use(
		cors.New(cors.Config{
			AllowOrigins:     "*",
			AllowMethods:     "GET,POST,OPTIONS,DELETE,PUT",
			AllowHeaders:     "origin, x-requested-with, content-type, authorization",
			AllowCredentials: true,
		}),
		recover.New(),
		contextualLoggerMiddleware,
		errorMiddleware,
		httpRequestLoggerMiddleware,
		authMiddleware,
	)
	base := instance.App.Group("/v1")
	base.Get("/ping", h.Ping)
	base.Post("/token", h.Token)
	return &instance
}
