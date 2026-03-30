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
}

type TokenRepository interface {
    SaveToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error
}

type ProductRepository interface {
    GetProducts(ctx context.Context) ([]model.Product, error)
    GetProductByID(ctx context.Context, id int64) (model.Product, error)
}

type CartRepository interface {
    GetCart(ctx context.Context, userID int64) (model.Cart, error)
    AddToCart(ctx context.Context, userID, productID int64, quantity int) error
}

type OrderRepository interface {
    CreateOrder(ctx context.Context, userID int64, address string) (model.Order, error)
    GetOrderByID(ctx context.Context, id int64) (model.Order, error)
    GetOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error)
}
