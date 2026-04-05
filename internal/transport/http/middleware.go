package http

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	stderrors "errors"
	appErrors "task5/internal/errors"

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
			return appErrors.Unauthorized("invalid token", err)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			return appErrors.Unauthorized("invalid token", nil)
		}

		userIDRaw := claims["sub"]
		userIDStr, ok := userIDRaw.(string)
		if !ok {
			return appErrors.Unauthorized("invalid token", nil)
		}

		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			return appErrors.Unauthorized("invalid token", err)
		}

		permissions, err := usecase.GetUserPermissions(ctx.Context(), userID)
		if err != nil {
			return appErrors.Forbidden("forbidden", err)
		}

		requiredPermission := buildPermission(ctx.Method(), ctx.Path())
		if !hasPermission(permissions, requiredPermission) {
			return appErrors.Forbidden("forbidden", nil)
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
	if err == nil {
		return nil
	}

	var appErr *appErrors.Error
	if !stderrors.As(err, &appErr) {
		appErr = appErrors.Internal(err)
	}

	fields := []zap.Field{
		zap.String("code", string(appErr.Code)),
		zap.String("message", appErr.Message),
		zap.Int("http_status", appErr.HTTPStatus),
	}

	if appErr.Err != nil {
		fields = append(fields, zap.Error(appErr.Err))
	}

	if len(appErr.Stack) > 0 {
		fields = append(fields, zap.Any("stack", appErr.Stack))
	}

	logger.Gist(ctx.Context()).Error("error occurred while request handling", fields...)

	return ctx.Status(appErr.HTTPStatus).JSON(fiber.Map{
		"code":    appErr.Code,
		"message": appErr.Message,
	})
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
