package model

import "time"

type OrderItem struct {
	ID              int64
	OrderID         int64
	ProductID       int64
	Quantity        int
	Price           float64
	DiscountPercent float64
	Product         *Product
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type Order struct {
	ID         int64
	UserID     int64
	Status     string
	TotalPrice float64
	Address    string
	Items      []OrderItem
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}