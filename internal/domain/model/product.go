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