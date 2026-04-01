package domain

import (
	"context"
	"fmt"
	"time"

	"task5/internal/domain/model"
)

type Usecase struct {
	userRepo    UserRepository
	tokenRepo   TokenRepository
	productRepo ProductRepository
	cartRepo    CartRepository
	orderRepo   OrderRepository
}

func NewUsecase(
	userRepo UserRepository,
	tokenRepo TokenRepository,
	productRepo ProductRepository,
	cartRepo CartRepository,
	orderRepo OrderRepository,
) *Usecase {
	return &Usecase{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		productRepo: productRepo,
		cartRepo:    cartRepo,
		orderRepo:   orderRepo,
	}
}

func (u *Usecase) Register(ctx context.Context, user model.User) (model.User, error) {
	// 1. проверка — существует ли пользователь
	existingUser, err := u.userRepo.GetUserByUsername(ctx, user.Username)
	if err == nil && existingUser.ID != 0 {
		return model.User{}, fmt.Errorf("user already exists")
	}

	// 2. TODO: хеширование пароля

	// 3. сохраняем пользователя
	createdUser, err := u.userRepo.CreateUser(ctx, user)
	if err != nil {
		return model.User{}, err
	}

	return createdUser, nil
}

func (u *Usecase) Login(ctx context.Context, username, password string) (string, error) {
	// 1. ищем пользователя
	user, err := u.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	// 2. TODO: проверка пароля

	// 3. генерируем токен (пока простой)
	token := fmt.Sprintf("token-%d", user.ID)

	// 4. сохраняем токен
	err = u.tokenRepo.SaveToken(ctx, user.ID, token, time.Now().Add(24*time.Hour))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (u *Usecase) GetUserByID(ctx context.Context, id int64) (model.User, error) {
	user, err := u.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

