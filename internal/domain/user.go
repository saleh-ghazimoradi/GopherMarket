package domain

import (
	"gorm.io/gorm"
	"time"
)

type UserRole string

const (
	Customer UserRole = "customer"
	Admin    UserRole = "admin"
)

type User struct {
	Id        uint   `gorm:"primaryKey"`
	Email     string `gorm:"uniqueIndex;not null"`
	Password  string `gorm:"not null"`
	FirstName string `gorm:"not null"`
	LastName  string `gorm:"not null"`
	Phone     string
	IsActive  bool     `gorm:"default:true"`
	Role      UserRole `gorm:"default:customer"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	RefreshTokens []RefreshToken
	Orders        []Order
	Cart          Cart
}

type RefreshToken struct {
	Id        uint      `gorm:"primaryKey"`
	UserId    uint      `gorm:"not null"`
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	User User
}
