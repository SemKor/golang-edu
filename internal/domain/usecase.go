package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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
	existingUser, err := u.userRepo.GetUserByUsername(ctx, user.Username)
	if err == nil && existingUser.ID != 0 {
		return model.User{}, fmt.Errorf("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, fmt.Errorf("cannot hash password: %w", err)
	}

	user.PasswordHash = string(hashedPassword)

	createdUser, err := u.userRepo.CreateUser(ctx, user)
	if err != nil {
		return model.User{}, err
	}

	err = u.userRepo.AssignRoleToUser(ctx, createdUser.ID, "user")
	if err != nil {
		return model.User{}, fmt.Errorf("cannot assign default role: %w", err)
	}

	return createdUser, nil
}

func (u *Usecase) Login(ctx context.Context, username, password string) (string, error) {
	user, err := u.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	claims := jwt.MapClaims{
		"sub": fmt.Sprintf("%d", user.ID),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}

	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	token, err := tokenObj.SignedString([]byte("00000000"))
	if err != nil {
		return "", err
	}

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

func (u *Usecase) GetUserPermissions(ctx context.Context, userID int64) ([]string, error) {
	return u.userRepo.GetUserPermissions(ctx, userID)
}

func (u *Usecase) GetProducts(ctx context.Context, filter model.ProductFilter) ([]model.Product, int, error) {
	return u.productRepo.GetProducts(ctx, filter)
}

func (u *Usecase) GetProductByID(ctx context.Context, id int64) (model.Product, error) {
	return u.productRepo.GetProductByID(ctx, id)
}

func (u *Usecase) GetCart(ctx context.Context, userID int64) (model.Cart, error) {
	return u.cartRepo.GetCart(ctx, userID)
}

func (u *Usecase) CreateOrder(ctx context.Context, userID int64, address string) (model.Order, error) {
	return u.orderRepo.CreateOrder(ctx, userID, address)
}

func (u *Usecase) GetOrderByID(ctx context.Context, id int64) (model.Order, error) {
	return u.orderRepo.GetOrderByID(ctx, id)
}

func (u *Usecase) GetOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error) {
	return u.orderRepo.GetOrdersByUser(ctx, userID)
}

func (u *Usecase) PayOrder(ctx context.Context, orderID int64) error {
	return u.orderRepo.PayOrder(ctx, orderID)
}

func (u *Usecase) GetProductDiscount(ctx context.Context, productID int64, isPremium bool) (float64, error) {
	return u.productRepo.GetProductDiscount(ctx, productID, isPremium)
}

func (u *Usecase) CancelOrder(ctx context.Context, orderID int64, userID int64) error {
	return u.orderRepo.CancelOrder(ctx, orderID, userID)
}

func (u *Usecase) GetOrderByIDForUser(ctx context.Context, orderID int64, userID int64) (model.Order, error) {
	return u.orderRepo.GetOrderByIDForUser(ctx, orderID, userID)
}
func (u *Usecase) ActivatePremium(ctx context.Context, userID int64) error {
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	return u.userRepo.ActivatePremium(ctx, userID, expiresAt)
}

func (u *Usecase) CreateProducts(ctx context.Context, products []model.ProductCreateInput) ([]model.Product, error) {
	return u.productRepo.CreateProducts(ctx, products)
}

func (u *Usecase) UpdateProducts(ctx context.Context, products []model.ProductUpdateInput) ([]model.Product, error) {
	return u.productRepo.UpdateProducts(ctx, products)
}

func (u *Usecase) DeleteProduct(ctx context.Context, id int64) error {
	return u.productRepo.DeleteProduct(ctx, id)
}

func (u *Usecase) ReplaceCart(ctx context.Context, userID int64, items []model.CartUpdateItem) error {
	return u.cartRepo.ReplaceCart(ctx, userID, items)
}




