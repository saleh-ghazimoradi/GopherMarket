package repository

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type ResetTokenRepository interface {
	Store(ctx context.Context, token string, userId uint, ttl time.Duration) error
	VerifyAndDelete(ctx context.Context, token string) (uint, error)
	Delete(ctx context.Context, token string) error
}

type resetTokenRepository struct {
	client *redis.Client
}

func (r *resetTokenRepository) Store(ctx context.Context, token string, userId uint, ttl time.Duration) error {
	key := "password_reset:" + token
	return r.client.Set(ctx, key, userId, ttl).Err()
}

func (r *resetTokenRepository) VerifyAndDelete(ctx context.Context, token string) (uint, error) {
	key := "password_reset:" + token

	userId, err := r.client.Get(ctx, key).Uint64()
	if err != nil {
		return 0, fmt.Errorf("invalid or expired token")
	}

	r.client.Del(ctx, key)
	return uint(userId), nil
}

func (r *resetTokenRepository) Delete(ctx context.Context, token string) error {
	key := "password_reset:" + token
	return r.client.Del(ctx, key).Err()
}

func NewResetTokenRepository(client *redis.Client) ResetTokenRepository {
	return &resetTokenRepository{
		client: client,
	}
}
