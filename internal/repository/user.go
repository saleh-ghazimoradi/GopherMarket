package repository

import (
	"context"
	"errors"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserById(ctx context.Context, id uint) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByEmailAndActive(ctx context.Context, email string, isActive bool) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	WithTx(tx *gorm.DB) UserRepository
}

type userRepository struct {
	dbWrite *gorm.DB
	dbRead  *gorm.DB
}

func (u *userRepository) CreateUser(ctx context.Context, user *domain.User) error {
	return u.dbWrite.WithContext(ctx).Create(user).Error
}

func (u *userRepository) GetUserById(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	if err := u.dbRead.WithContext(ctx).First(&user, id).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrsNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (u *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := u.dbRead.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrsNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (u *userRepository) GetUserByEmailAndActive(ctx context.Context, email string, active bool) (*domain.User, error) {
	var user domain.User
	if err := u.dbRead.WithContext(ctx).Where("email = ? AND is_active = ?", email, active).First(&user).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrsNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (u *userRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	return u.dbWrite.WithContext(ctx).Save(user).Error
}

func (u *userRepository) WithTx(tx *gorm.DB) UserRepository {
	return &userRepository{
		dbWrite: tx,
		dbRead:  tx,
	}
}

func NewUserRepository(dbWrite, dbRead *gorm.DB) UserRepository {
	return &userRepository{
		dbWrite: dbWrite,
		dbRead:  dbRead,
	}
}
