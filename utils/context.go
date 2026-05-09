package utils

import "context"

type ContextKey string

const (
	UserIdKey ContextKey = "user_id"
	EmailKey  ContextKey = "email"
	RoleKey   ContextKey = "role"
)

func WithUserId(ctx context.Context, id uint) context.Context {
	return context.WithValue(ctx, UserIdKey, id)
}

func UserIdFromContext(ctx context.Context) (uint, bool) {
	id, ok := ctx.Value(UserIdKey).(uint)
	return id, ok
}

func WithEmailKey(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, EmailKey, email)
}

func EmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(EmailKey).(string)
	return email, ok
}

func WithRoleKey(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, RoleKey, role)
}

func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(RoleKey).(string)
	return role, ok
}
