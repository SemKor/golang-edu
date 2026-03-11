package http

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"task5/internal/logger"
)

func authMiddleware(ctx *fiber.Ctx) error {
	if ctx.Path() == "/v1/register" ||
		ctx.Path() == "/v1/token" {
		return ctx.Next()
	}
	tokenRaw := strings.ReplaceAll(ctx.Get("authorization"), "Bearer ", "")
	token, err := jwt.Parse(tokenRaw, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("00000000"), nil
	})
	if err != nil {
		ctx.Status(http.StatusUnauthorized)
		return fmt.Errorf("cannot parse token: %w", err)
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println(claims["sub"]) // Это и есть идентификатор пользователя который мы проставили в токене при его получении.
		// TODO Далее имея идентификатор пользователя, мы должны его авторизовать, запросив его роли и права
	} else {
		ctx.Status(http.StatusUnauthorized)
		return fmt.Errorf("cannot parse token: %w", err)
	}
	return ctx.Next()
}

func errorMiddleware(ctx *fiber.Ctx) error {
	err := ctx.Next()
	if err != nil {
		logger.Gist(ctx.Context()).Error("error occurred while request handling", zap.Error(err))
	}
	return err
}

func contextualLoggerMiddleware(c *fiber.Ctx) error {
	traceId := uuid.NewString()
	lg := logger.Gist(c.Context())
	lg = lg.With(zap.String("trace-id", traceId))
	logger.SetLogger(lg, c.Context().SetUserValue)
	return c.Next()
}

func httpRequestLoggerMiddleware(c *fiber.Ctx) error {
	lg := logger.Gist(c.Context())
	start := time.Now()
	path := string(c.Request().URI().Path())
	raw := string(c.Request().URI().QueryString())
	if raw != "" {
		path = path + "?" + raw
	}
	defer func() {
		lg.With(
			zap.Int("status", c.Response().StatusCode()),
			zap.String("duration", fmt.Sprintf("%v", time.Now().Sub(start))),
			zap.String("client-ip", c.IP()),
			zap.String("method", c.Method()),
			zap.String("path", path),
		).Debug("api call")
	}()
	return c.Next()
}
