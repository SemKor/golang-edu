package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"task5/internal/domain/model"
)

type Repository struct {
	mu          sync.RWMutex
	users       map[int64]model.User
	usersByName map[string]int64
	tokens      map[string]int64
	nextUserID  int64
}

func NewRepository() *Repository {
	return &Repository{
		users:       make(map[int64]model.User),
		usersByName: make(map[string]int64),
		tokens:      make(map[string]int64),
		nextUserID:  1,
	}
}

func (r *Repository) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.usersByName[user.Username]; exists {
		return model.User{}, fmt.Errorf("user already exists")
	}

	user.ID = r.nextUserID
	r.nextUserID++

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	r.users[user.ID] = user
	r.usersByName[user.Username] = user.ID

	return user, nil
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.usersByName[username]
	if !exists {
		return model.User{}, fmt.Errorf("user not found")
	}

	user, exists := r.users[id]
	if !exists {
		return model.User{}, fmt.Errorf("user not found")
	}

	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return model.User{}, fmt.Errorf("user not found")
	}

	return user, nil
}

func (r *Repository) SaveToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tokens[token] = userID
	return nil
}

func (r *Repository) GetProducts(ctx context.Context) ([]model.Product, error) {
	return []model.Product{}, nil
}

func (r *Repository) GetProductByID(ctx context.Context, id int64) (model.Product, error) {
	return model.Product{}, fmt.Errorf("product not found")
}

func (r *Repository) GetCart(ctx context.Context, userID int64) (model.Cart, error) {
	return model.Cart{}, fmt.Errorf("cart not found")
}

func (r *Repository) AddToCart(ctx context.Context, userID, productID int64, quantity int) error {
	return nil
}

func (r *Repository) CreateOrder(ctx context.Context, userID int64, address string) (model.Order, error) {
	return model.Order{}, fmt.Errorf("not implemented")
}

func (r *Repository) GetOrderByID(ctx context.Context, id int64) (model.Order, error) {
	return model.Order{}, fmt.Errorf("order not found")
}

func (r *Repository) GetOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error) {
	return []model.Order{}, nil
}