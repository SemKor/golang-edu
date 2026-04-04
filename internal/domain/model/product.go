package model

import "time"

type Product struct {
	ID          int64
	Name        string
	Description string
	Price       float64
	CategoryID  int64
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type ProductCreateInput struct {
	Name             string
	Description      string
	Price            float64
	CategoryID       int64
	PremiumDiscount  float64
	CategoryDiscount float64
}

type ProductUpdateInput struct {
	ID               int64
	Name             string
	Description      string
	Price            float64
	CategoryID       int64
	PremiumDiscount  float64
	CategoryDiscount float64
}

type ProductFilter struct {
	Name      string
	Category  []int64
	MinPrice  float64
	MaxPrice  float64
	Offset    int
	Count     int
}