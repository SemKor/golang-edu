package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"task5/internal/config"
	"task5/internal/domain"
)

type Server struct {
	App *fiber.App
}

// NewServer TODO закончить инициализацию сервера обработчиками
func NewServer(cfg *config.Config, usecase *domain.Usecase) *Server {
	instance := Server{
		App: fiber.New(),
	}
	h := newHandler(cfg, usecase)
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
		authMiddleware(usecase),
	)
	base := instance.App.Group("/v1")
	base.Get("/ping", h.Ping)
	base.Post("/token", h.Token)
	base.Post("/register", h.Register)
	base.Get("/user", h.GetUser)
	base.Get("/products", h.GetProducts)
	base.Get("/product/:id", h.GetProductByID)
	base.Put("/cart", h.ReplaceCart)
	base.Get("/cart", h.GetCart)
	base.Post("/order", h.CreateOrder)
	base.Get("/order/:id", h.GetOrder)
	base.Get("/orders", h.GetOrders)
	base.Post("/pay", h.PayOrder)
	base.Delete("/order/:id", h.CancelOrder)
	base.Post("/products", h.CreateProducts)
	base.Put("/products", h.UpdateProducts)
	base.Delete("/product/:id", h.DeleteProduct)
	return &instance
}
