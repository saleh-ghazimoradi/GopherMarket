package repository

import (
	"context"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"gorm.io/gorm"
	"time"
)

type TokenRepository interface {
	CreateToken(ctx context.Context, token *domain.RefreshToken) error
	GetValidRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	DeleteTokenById(ctx context.Context, id uint) error
	DeleteToken(ctx context.Context, token string) error
	WithTx(tx *gorm.DB) TokenRepository
}

type tokenRepository struct {
	dbWrite *gorm.DB
	dbRead  *gorm.DB
}

func (t *tokenRepository) CreateToken(ctx context.Context, token *domain.RefreshToken) error {
	return t.dbWrite.WithContext(ctx).Create(token).Error
}

func (t *tokenRepository) GetValidRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var refreshToken domain.RefreshToken
	if err := t.dbRead.WithContext(ctx).
		Where("token = ? AND expires_at > ?", token, time.Now()).
		First(&refreshToken).Error; err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (t *tokenRepository) DeleteTokenById(ctx context.Context, id uint) error {
	return t.dbWrite.WithContext(ctx).Delete(&domain.RefreshToken{}, id).Error
}

func (t *tokenRepository) DeleteToken(ctx context.Context, token string) error {
	return t.dbWrite.WithContext(ctx).
		Where("token = ?", token).
		Delete(&domain.RefreshToken{}).Error
}

func (t *tokenRepository) WithTx(tx *gorm.DB) TokenRepository {
	return &tokenRepository{
		dbWrite: tx,
		dbRead:  tx,
	}
}

func NewTokenRepository(dbWrite *gorm.DB, dbRead *gorm.DB) TokenRepository {
	return &tokenRepository{
		dbWrite: dbWrite,
		dbRead:  dbRead,
	}
}
