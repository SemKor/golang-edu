package http

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	

	"task5/internal/domain"
	"task5/internal/logger"
)

func authMiddleware(usecase *domain.Usecase) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if ctx.Path() == "/v1/register" ||
			ctx.Path() == "/v1/token" ||
			ctx.Path() == "/v1/ping" {
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

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			ctx.Status(http.StatusUnauthorized)
			return fmt.Errorf("invalid token")
		}

		userIDRaw := claims["sub"]
		userIDStr, ok := userIDRaw.(string)
		if !ok {
			ctx.Status(http.StatusUnauthorized)
			return fmt.Errorf("invalid token subject")
		}

		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			ctx.Status(http.StatusUnauthorized)
			return fmt.Errorf("invalid user id in token: %w", err)
		}

		permissions, err := usecase.GetUserPermissions(ctx.Context(), userID)
		if err != nil {
			ctx.Status(http.StatusForbidden)
			return fmt.Errorf("cannot get user permissions: %w", err)
		}

		requiredPermission := buildPermission(ctx.Method(), ctx.Path())
		if !hasPermission(permissions, requiredPermission) {
			ctx.Status(http.StatusForbidden)
			return fmt.Errorf("forbidden")
		}

		ctx.Locals("user_id", userIDStr)
		ctx.Locals("permissions", permissions)

		return ctx.Next()
	}
}

func buildPermission(method, path string) string {
	path = strings.TrimPrefix(path, "/v1")

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 1 {
		last := parts[len(parts)-1]
		if _, err := strconv.ParseInt(last, 10, 64); err == nil {
			parts = parts[:len(parts)-1]
		}
	}

	cleanPath := strings.Join(parts, "/")
	return method + "/" + cleanPath
}

func hasPermission(permissions []string, required string) bool {
	for _, permission := range permissions {
		if permission == required {
			return true
		}
	}
	return false
}

func errorMiddleware(ctx *fiber.Ctx) error {
	err := ctx.Next()
	if err != nil {
		logger.Gist(ctx.Context()).Error("error occurred while request handling", zap.Error(err))
		return nil
	}
	return nil
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
