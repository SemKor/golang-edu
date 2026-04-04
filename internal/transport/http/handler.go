package http

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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
		Login     string `json:"login"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"firstname"`
		LastName  string `json:"lastname"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if strings.TrimSpace(req.Login) == "" {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "login is required",
		})
	}

	if strings.TrimSpace(req.Email) == "" {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "email is required",
		})
	}

	if strings.TrimSpace(req.Password) == "" {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "password is required",
		})
	}

	if strings.TrimSpace(req.FirstName) == "" {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "firstname is required",
		})
	}

	if strings.TrimSpace(req.LastName) == "" {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "lastname is required",
		})
	}

	user := model.User{
		Username:     req.Login,
		Email:        req.Email,
		PasswordHash: req.Password,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	}

	createdUser, err := h.usecase.Register(ctx.Context(), user)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"id": createdUser.ID,
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
		"firstname":  user.FirstName,
		"lastname":   user.LastName,
		"hasPremium": user.IsPremium,
	})
}

func (h *handler) GetProducts(ctx *fiber.Ctx) error {
	userIDRaw := ctx.Locals("user_id")

	idStr, ok := userIDRaw.(string)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	user, err := h.usecase.GetUserByID(ctx.Context(), userID)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	var filter model.ProductFilter

	filter.Name = strings.TrimSpace(ctx.Query("name"))

	categoryParams := strings.TrimSpace(ctx.Query("category"))
	if categoryParams != "" {
		parts := strings.Split(categoryParams, ",")
		for _, p := range parts {
			id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
			if err != nil {
				return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
					"error": "invalid category",
				})
			}
			filter.Category = append(filter.Category, id)
		}
	}

	if v := strings.TrimSpace(ctx.Query("minPrice")); v != "" {
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid minPrice",
			})
		}
		filter.MinPrice = val
	}

	if v := strings.TrimSpace(ctx.Query("maxPrice")); v != "" {
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid maxPrice",
			})
		}
		filter.MaxPrice = val
	}

	var minDiscount *float64
	if v := strings.TrimSpace(ctx.Query("minDiscount")); v != "" {
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid minDiscount",
			})
		}
		minDiscount = &val
	}

	var maxDiscount *float64
	if v := strings.TrimSpace(ctx.Query("maxDiscount")); v != "" {
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid maxDiscount",
			})
		}
		maxDiscount = &val
	}

	offset := 0
	if v := strings.TrimSpace(ctx.Query("offset")); v != "" {
		val, err := strconv.Atoi(v)
		if err != nil || val < 0 {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid offset",
			})
		}
		offset = val
	}

	count := -1
	if v := strings.TrimSpace(ctx.Query("count")); v != "" {
		val, err := strconv.Atoi(v)
		if err != nil || val < 0 {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid count",
			})
		}
		count = val
	}

	// Важно:
	// offset/count здесь не передаем в repository, чтобы totalCount считался
	// после применения discount-фильтров, а пагинация была последним шагом.
	filter.Offset = 0
	filter.Count = 0

	products, _, err := h.usecase.GetProducts(ctx.Context(), filter)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "cannot get products",
		})
	}

	type productResponse struct {
		ID       int64   `json:"id"`
		Name     string  `json:"name"`
		Category int64   `json:"category"`
		Price    float64 `json:"price"`
		Discount float64 `json:"discount"`
	}

	result := make([]productResponse, 0, len(products))

	for _, product := range products {
		discount, err := h.usecase.GetProductDiscount(ctx.Context(), product.ID, user.IsPremium)
		if err != nil {
			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "cannot get product discount",
			})
		}

		if minDiscount != nil && discount < *minDiscount {
			continue
		}
		if maxDiscount != nil && discount > *maxDiscount {
			continue
		}

		finalPrice := product.Price * (1 - discount/100)

		result = append(result, productResponse{
			ID:       product.ID,
			Name:     product.Name,
			Category: product.CategoryID,
			Price:    finalPrice,
			Discount: discount,
		})
	}

	totalCount := len(result)

	if offset >= len(result) {
		result = []productResponse{}
	} else if offset > 0 {
		result = result[offset:]
	}

	if count >= 0 && count < len(result) {
		result = result[:count]
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"products":   result,
		"totalCount": totalCount,
	})
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

	userIDRaw := ctx.Locals("user_id")
	idStr, ok := userIDRaw.(string)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	user, err := h.usecase.GetUserByID(ctx.Context(), userID)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	discount, err := h.usecase.GetProductDiscount(ctx.Context(), product.ID, user.IsPremium)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "cannot get product discount",
		})
	}

	finalPrice := product.Price * (1 - discount/100)

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"id":          product.ID,
		"name":        product.Name,
		"description": product.Description,
		"category_id": product.CategoryID,
		"price":       finalPrice,
		"discount":    discount,
		"is_active":   product.IsActive,
		"created_at":  product.CreatedAt,
		"updated_at":  product.UpdatedAt,
	})
}

func (h *handler) ReplaceCart(ctx *fiber.Ctx) error {
	userIDRaw := ctx.Locals("user_id")

	idStr, ok := userIDRaw.(string)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	var req struct {
		Products []struct {
			ID       int64 `json:"id"`
			Quantity int   `json:"quantity"`
		} `json:"products"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid body",
		})
	}

	items := make([]model.CartUpdateItem, 0, len(req.Products))
	for _, product := range req.Products {
		if product.ID <= 0 {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid product id",
			})
		}
		if product.Quantity <= 0 {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid quantity",
			})
		}

		items = append(items, model.CartUpdateItem{
			ProductID: product.ID,
			Quantity:  product.Quantity,
		})
	}

	err = h.usecase.ReplaceCart(ctx.Context(), userID, items)
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

	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	cart, err := h.usecase.GetCart(ctx.Context(), userID)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "cart not found",
		})
	}

	type cartProductResponse struct {
		ID       int64   `json:"id"`
		Quantity int     `json:"quantity"`
		Price    float64 `json:"price"`
		Discount float64 `json:"discount"`
		Amount   float64 `json:"amount"`
	}

	products := make([]cartProductResponse, 0, len(cart.Items))
	var totalAmount float64

	for _, item := range cart.Items {
		if item.Product == nil {
			continue
		}

		price := item.Product.Price
		amount := price * float64(item.Quantity)

		products = append(products, cartProductResponse{
			ID:       item.ProductID,
			Quantity: item.Quantity,
			Price:    price,
			Discount: item.DiscountPercent,
			Amount:   amount,
		})

		totalAmount += amount
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"products":    products,
		"totalAmount": totalAmount,
	})
}

func (h *handler) CreateOrder(ctx *fiber.Ctx) error {
	userIDRaw := ctx.Locals("user_id")

	idStr, ok := userIDRaw.(string)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	var req struct {
		Address string `json:"address"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid body",
		})
	}

	if strings.TrimSpace(req.Address) == "" {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "address is required",
		})
	}

	order, err := h.usecase.CreateOrder(ctx.Context(), userID, req.Address)
	if err != nil {
		switch err.Error() {
		case "cart not found", "cart is empty":
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		default:
			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"id": order.ID,
	})
}

func (h *handler) GetOrder(ctx *fiber.Ctx) error {
	userIDRaw := ctx.Locals("user_id")

	idStr, ok := userIDRaw.(string)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	orderID, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	order, err := h.usecase.GetOrderByIDForUser(ctx.Context(), orderID, userID)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "order not found",
		})
	}

	type orderProductResponse struct {
		ID       int64   `json:"id"`
		Quantity int     `json:"quantity"`
		Price    float64 `json:"price"`
		Discount float64 `json:"discount"`
		Amount   float64 `json:"amount"`
	}

	products := make([]orderProductResponse, 0, len(order.Items))
	var totalAmount float64

	for _, item := range order.Items {
		finalPrice := item.Price * (1 - item.DiscountPercent/100)
		amount := finalPrice * float64(item.Quantity)

		products = append(products, orderProductResponse{
			ID:       item.ProductID,
			Quantity: item.Quantity,
			Price:    finalPrice,
			Discount: item.DiscountPercent,
			Amount:   amount,
		})

		totalAmount += amount
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"id":          order.ID,
		"createdAt":   order.CreatedAt,
		"products":    products,
		"totalAmount": totalAmount,
	})
}

func (h *handler) GetOrders(ctx *fiber.Ctx) error {
	userIDRaw := ctx.Locals("user_id")

	idStr, ok := userIDRaw.(string)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	orders, err := h.usecase.GetOrdersByUser(ctx.Context(), userID)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "cannot get orders",
		})
	}

	type orderResponse struct {
		ID        int64     `json:"id"`
		CreatedAt time.Time `json:"createdAt"`
	}

	result := make([]orderResponse, 0, len(orders))
	for _, order := range orders {
		result = append(result, orderResponse{
			ID:        order.ID,
			CreatedAt: order.CreatedAt,
		})
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"orders": result,
	})
}

func (h *handler) PayOrder(ctx *fiber.Ctx) error {
	userIDRaw := ctx.Locals("user_id")

	idStr, ok := userIDRaw.(string)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	var req struct {
		PaymentType string  `json:"paymentType"`
		Amount      float64 `json:"amount"`
		OrderID     int64   `json:"order_id"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid body",
		})
	}

	if req.Amount < 0 {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid amount",
		})
	}

	switch req.PaymentType {
	case "order":
		if req.OrderID <= 0 {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid order_id",
			})
		}

		order, err := h.usecase.GetOrderByIDForUser(ctx.Context(), req.OrderID, userID)
		if err != nil {
			return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "order not found",
			})
		}

		if order.Status == "cancelled" {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "cancelled order cannot be paid",
			})
		}

		if order.Status == "paid" {
			return ctx.SendStatus(http.StatusOK)
		}

		if req.Amount != order.TotalPrice {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid payment amount",
			})
		}

		err = h.usecase.PayOrder(ctx.Context(), req.OrderID)
		if err != nil {
			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return ctx.SendStatus(http.StatusOK)

	case "premium":
		if req.Amount <= 0 {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid amount",
			})
		}

		err := h.usecase.ActivatePremium(ctx.Context(), userID)
		if err != nil {
			if err.Error() == "user not found" {
				return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
					"error": err.Error(),
				})
			}

			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return ctx.SendStatus(http.StatusOK)

	default:
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid paymentType",
		})
	}
}

func (h *handler) CancelOrder(ctx *fiber.Ctx) error {
	userIDRaw := ctx.Locals("user_id")

	idStr, ok := userIDRaw.(string)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	orderID, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid order id",
		})
	}

	err = h.usecase.CancelOrder(ctx.Context(), orderID, userID)
	if err != nil {
		switch err.Error() {
		case "order not found":
			return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		case "paid order cannot be cancelled":
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		default:
			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	return ctx.SendStatus(http.StatusOK)
}

func (h *handler) CreateProducts(ctx *fiber.Ctx) error {
	var req []struct {
		Name             string  `json:"name"`
		Description      string  `json:"description"`
		Price            float64 `json:"price"`
		CategoryID       int64   `json:"categoryId"`
		PremiumDiscount  float64 `json:"premiumDiscount"`
		CategoryDiscount float64 `json:"categoryDiscount"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if len(req) == 0 {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "empty products list",
		})
	}

	products := make([]model.ProductCreateInput, 0, len(req))
	for _, item := range req {
		if strings.TrimSpace(item.Name) == "" {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "product name is required",
			})
		}
		if item.Price < 0 {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "product price must be non-negative",
			})
		}
		if item.CategoryID <= 0 {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "categoryId must be positive",
			})
		}
		if item.PremiumDiscount < 0 || item.PremiumDiscount > 100 {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "premiumDiscount must be between 0 and 100",
			})
		}
		if item.CategoryDiscount < 0 || item.CategoryDiscount > 100 {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "categoryDiscount must be between 0 and 100",
			})
		}

		products = append(products, model.ProductCreateInput{
			Name:             item.Name,
			Description:      item.Description,
			Price:            item.Price,
			CategoryID:       item.CategoryID,
			PremiumDiscount:  item.PremiumDiscount,
			CategoryDiscount: item.CategoryDiscount,
		})
	}

	created, err := h.usecase.CreateProducts(ctx.Context(), products)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(http.StatusCreated).JSON(fiber.Map{
		"products": created,
	})
}

func (h *handler) UpdateProducts(ctx *fiber.Ctx) error {
	var req []struct {
		ID               int64   `json:"id"`
		Name             string  `json:"name"`
		Description      string  `json:"description"`
		Price            float64 `json:"price"`
		CategoryID       int64   `json:"categoryId"`
		PremiumDiscount  float64 `json:"premiumDiscount"`
		CategoryDiscount float64 `json:"categoryDiscount"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if len(req) == 0 {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "empty products list",
		})
	}

	products := make([]model.ProductUpdateInput, 0, len(req))
	for _, item := range req {
		if item.ID <= 0 {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "product id must be positive",
			})
		}
		if strings.TrimSpace(item.Name) == "" {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "product name is required",
			})
		}
		if item.Price < 0 {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "product price must be non-negative",
			})
		}
		if item.CategoryID <= 0 {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "categoryId must be positive",
			})
		}
		if item.PremiumDiscount < 0 || item.PremiumDiscount > 100 {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "premiumDiscount must be between 0 and 100",
			})
		}
		if item.CategoryDiscount < 0 || item.CategoryDiscount > 100 {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "categoryDiscount must be between 0 and 100",
			})
		}

		products = append(products, model.ProductUpdateInput{
			ID:               item.ID,
			Name:             item.Name,
			Description:      item.Description,
			Price:            item.Price,
			CategoryID:       item.CategoryID,
			PremiumDiscount:  item.PremiumDiscount,
			CategoryDiscount: item.CategoryDiscount,
		})
	}

	updated, err := h.usecase.UpdateProducts(ctx.Context(), products)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"products": updated,
	})
}

func (h *handler) DeleteProduct(ctx *fiber.Ctx) error {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product id",
		})
	}

	err = h.usecase.DeleteProduct(ctx.Context(), id)
	if err != nil {
		if err.Error() == "product not found" {
			return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.SendStatus(http.StatusOK)
}



