package model

import "time"

type User struct {
	ID               int64
	Username         string
	Email            string
	PasswordHash     string
	FirstName        string
	LastName         string
	IsPremium        bool
	PremiumExpiresAt *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}
