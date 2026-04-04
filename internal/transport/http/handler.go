package http

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
	"strconv"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	

	"task5/internal/config"
	"task5/internal/domain"
	"task5/internal/domain/model"

)

const (
	clientId     = "000000"
	secret       = "999999"
	clientDomain = "http://localhost"
)

// handler
// TODO инициализировать доменной структурой (domain) в которой реализована бизнес-логика приложения.
// TODO Реаоизовать методы соглагно контрактам, в методах вызывать соответствующие методы доменной структуры для получения результата.
type handler struct {
	cfg     *config.Config
	manager *manage.Manager
	usecase *domain.Usecase
}

func newHandler(cfg *config.Config, usecase *domain.Usecase) *handler {
	manager := manage.NewDefaultManager()
	manager.MustTokenStorage(store.NewMemoryTokenStore())
	clientStore := store.NewClientStore()
	clientStore.Set(clientId, &models.Client{
		ID:     clientId,
		Secret: secret,
		Domain: clientDomain,
	})
	manager.MapClientStorage(clientStore)
	return &handler{
		manager: manager,
		cfg:     cfg,
		usecase: usecase,
	}
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
			UserID:         "1",
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
		return ctx.Status(http.StatusOK).JSON(data)
	}
	// TODO дописать логику для генерации токенов для зарегистрированных пользователей и сохранения их в БД
	// если пользователь есть в БД то генерируем токен и сохраняем его в БД

	token, err := h.usecase.Login(ctx.Context(), username, password)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"access_token": token,
		"token_type":   "Bearer",
	})

}

func (h *handler) Register(ctx *fiber.Ctx) error {
	var req struct {
		Username  string `json:"username"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	user := model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.Password, // пока временно так, потом заменим на hash
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	}

	createdUser, err := h.usecase.Register(ctx.Context(), user)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(http.StatusCreated).JSON(fiber.Map{
		"id":         createdUser.ID,
		"username":   createdUser.Username,
		"email":      createdUser.Email,
		"first_name": createdUser.FirstName,
		"last_name":  createdUser.LastName,
	})
}

func (h *handler) GetUser(ctx *fiber.Ctx) error {
	userIDRaw := ctx.Locals("user_id")

	idStr, ok := userIDRaw.(string)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	user, err := h.usecase.GetUserByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"is_premium": user.IsPremium,
	})
}

func (h *handler) GetProducts(ctx *fiber.Ctx) error {
	products, err := h.usecase.GetProducts(ctx.Context())
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "cannot get products",
		})
	}

	return ctx.Status(http.StatusOK).JSON(products)
}

func (h *handler) GetProductByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product id",
		})
	}

	product, err := h.usecase.GetProductByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "product not found",
		})
	}

	return ctx.Status(http.StatusOK).JSON(product)
}

func (h *handler) AddToCart(ctx *fiber.Ctx) error {
	userIDRaw := ctx.Locals("user_id")

	idStr, ok := userIDRaw.(string)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	userID, _ := strconv.ParseInt(idStr, 10, 64)

	var req struct {
		ProductID int64 `json:"product_id"`
		Quantity  int   `json:"quantity"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid body",
		})
	}

	err := h.usecase.AddToCart(ctx.Context(), userID, req.ProductID, req.Quantity)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.SendStatus(http.StatusOK)
}

func (h *handler) GetCart(ctx *fiber.Ctx) error {
	userIDRaw := ctx.Locals("user_id")

	idStr, ok := userIDRaw.(string)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	userID, _ := strconv.ParseInt(idStr, 10, 64)

	cart, err := h.usecase.GetCart(ctx.Context(), userID)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "cart not found",
		})
	}

	return ctx.Status(http.StatusOK).JSON(cart)
}

func (h *handler) CreateOrder(ctx *fiber.Ctx) error {
	userIDRaw := ctx.Locals("user_id")

	idStr, ok := userIDRaw.(string)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	userID, _ := strconv.ParseInt(idStr, 10, 64)

	var req struct {
		Address string `json:"address"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid body",
		})
	}

	order, err := h.usecase.CreateOrder(ctx.Context(), userID, req.Address)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(http.StatusCreated).JSON(order)
}

func (h *handler) GetOrder(ctx *fiber.Ctx) error {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	order, err := h.usecase.GetOrderByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "order not found",
		})
	}

	return ctx.Status(http.StatusOK).JSON(order)
}

func (h *handler) GetOrders(ctx *fiber.Ctx) error {
	userIDRaw := ctx.Locals("user_id")

	idStr, ok := userIDRaw.(string)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	userID, _ := strconv.ParseInt(idStr, 10, 64)

	orders, err := h.usecase.GetOrdersByUser(ctx.Context(), userID)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "cannot get orders",
		})
	}

	return ctx.Status(http.StatusOK).JSON(orders)
}

func (h *handler) PayOrder(ctx *fiber.Ctx) error {
	var req struct {
		OrderID int64 `json:"order_id"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid body",
		})
	}

	if req.OrderID <= 0 {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid order_id",
		})
	}

	err := h.usecase.PayOrder(ctx.Context(), req.OrderID)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.SendStatus(http.StatusOK)
}