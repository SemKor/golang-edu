package domain

import (
	"context"

	"task5/internal/domain/model"
)

// TODO интерфейсы, за которыми мы закрывамся от зависимостей на другие слои

// UserRepository пример
type UserRepository interface {
	GetUser(ctx context.Context) model.User
}
