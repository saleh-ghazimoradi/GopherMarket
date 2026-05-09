package domain

import (
	"gorm.io/gorm"
	"time"
)

type RefreshToken struct {
	Id        uint      `gorm:"primaryKey"`
	UserId    uint      `gorm:"not null"`
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	User User
}
