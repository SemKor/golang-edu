package domain

import (
	"context"
	"time"

	"task5/internal/domain/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user model.User) (model.User, error)
	GetUserByUsername(ctx context.Context, username string) (model.User, error)
	GetUserByID(ctx context.Context, id int64) (model.User, error)
	GetUserRoles(ctx context.Context, userID int64) ([]string, error)
	GetUserPermissions(ctx context.Context, userID int64) ([]string, error)
	AssignRoleToUser(ctx context.Context, userID int64, roleName string) error
	ActivatePremium(ctx context.Context, userID int64, expiresAt time.Time) error
}

type TokenRepository interface {
	SaveToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error
}

type ProductRepository interface {
	GetProducts(ctx context.Context, filter model.ProductFilter) ([]model.Product, int, error)
	GetProductByID(ctx context.Context, id int64) (model.Product, error)
	GetProductDiscount(ctx context.Context, productID int64, isPremium bool) (float64, error)
	CreateProducts(ctx context.Context, products []model.ProductCreateInput) ([]model.Product, error)
	UpdateProducts(ctx context.Context, products []model.ProductUpdateInput) ([]model.Product, error)
	DeleteProduct(ctx context.Context, id int64) error
}

type CartRepository interface {
	GetCart(ctx context.Context, userID int64) (model.Cart, error)
	ReplaceCart(ctx context.Context, userID int64, items []model.CartUpdateItem) error
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, userID int64, address string) (model.Order, error)
	GetOrderByID(ctx context.Context, id int64) (model.Order, error)
	GetOrderByIDForUser(ctx context.Context, orderID int64, userID int64) (model.Order, error)
	GetOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error)
	PayOrder(ctx context.Context, orderID int64) error
	CancelOrder(ctx context.Context, orderID int64, userID int64) error
}
