package http

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"

	"task5/internal/config"
)

const (
	clientId = "000000"
	secret   = "999999"
	domain   = "http://localhost"
)

// handler
// TODO инициализировать доменной структурой (domain) в которой реализована бизнес-логика приложения.
// TODO Реаоизовать методы соглагно контрактам, в методах вызывать соответствующие методы доменной структуры для получения результата.
type handler struct {
	cfg     *config.Config
	manager *manage.Manager
}

func newHandler(cfg *config.Config) *handler {
	manager := manage.NewDefaultManager()
	manager.MustTokenStorage(store.NewMemoryTokenStore())
	clientStore := store.NewClientStore()
	clientStore.Set(clientId, &models.Client{
		ID:     clientId,
		Secret: secret,
		Domain: domain,
	})
	manager.MapClientStorage(clientStore)
	return &handler{manager: manager, cfg: cfg}
}

func (h *handler) Ping(ctx *fiber.Ctx) error {
	ctx.Status(http.StatusOK)
	return nil
}

func (h *handler) Token(ctx *fiber.Ctx) error {
	params, err := url.ParseQuery(string(ctx.Body()))
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return fmt.Errorf("cannot parse body: %w", err)
	}
	if params.Get("grant_type") != oauth2.PasswordCredentials.String() {
		ctx.Status(http.StatusBadRequest)
		return fmt.Errorf("invalid grant type")
	}
	username := params.Get("username")
	password := params.Get("password")
	if h.cfg.TestUser.Allowed &&
		username == h.cfg.TestUser.Name &&
		password == h.cfg.TestUser.Password {
		h.manager.MapAccessGenerate(generates.NewJWTAccessGenerate("", []byte("00000000"), jwt.SigningMethodHS512))
		ti, err := h.manager.GenerateAccessToken(ctx.Context(), oauth2.PasswordCredentials, &oauth2.TokenGenerateRequest{
			ClientID:       clientId,
			ClientSecret:   secret,
			UserID:         "test",
			Scope:          "[read, write]",
			AccessTokenExp: 5 * time.Minute,
		})
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			return fmt.Errorf("cannot generate test token: %w", err)
		}
		data := map[string]interface{}{
			"access_token":  ti.GetAccess(),
			"refresh_token": ti.GetRefresh(),
			"token_type":    "Bearer",
			"expires_in":    int64(ti.GetAccessExpiresIn() / time.Second),
		}
		ctx.Set("Content-Type", "application/json;charset=UTF-8")
		ctx.Set("Cache-Control", "no-store")
		ctx.Set("Pragma", "no-cache")
		ctx.JSON(data)
		ctx.Status(http.StatusOK)
	}
	// TODO дописать логику для генерации токенов для зарегистрированных пользователей и сохранения их в БД
	// если пользователь есть в БД то генерируем токен и сохраняем его в БД
	ctx.Status(http.StatusOK)
	return nil
}
