package model

import "time"

type CartUpdateItem struct {
	ProductID int64
	Quantity  int
}

type CartItem struct {
	ID              int64
	CartID          int64
	ProductID       int64
	Quantity        int
	DiscountPercent float64
	Product         *Product
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type Cart struct {
	ID        int64
	UserID    int64
	Items     []CartItem
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}