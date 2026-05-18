package service

import (
	"context"
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type UserService interface {
	GetUserById(ctx context.Context, id uint) (*dto.UserResponse, error)
	GetUserProfile(ctx context.Context, id uint) (*dto.UserResponse, error)
	UpdateUserProfile(ctx context.Context, id uint, req *dto.UpdateProfileRequest) (*dto.UserResponse, error)
}

type userService struct {
	userRepository repository.UserRepository
	tracer         trace.Tracer
}

func (u *userService) GetUserById(ctx context.Context, id uint) (*dto.UserResponse, error) {
	ctx, span := u.tracer.Start(ctx, "UserService.GetUserById",
		trace.WithAttributes(attribute.Int64("user.id", int64(id))))
	defer span.End()

	user, err := u.userRepository.GetUserById(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "user not found")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return u.toUserResp(user), nil
}

func (u *userService) GetUserProfile(ctx context.Context, id uint) (*dto.UserResponse, error) {
	return u.GetUserById(ctx, id)
}

func (u *userService) UpdateUserProfile(ctx context.Context, id uint, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	ctx, span := u.tracer.Start(ctx, "UserService.UpdateUserProfile",
		trace.WithAttributes(attribute.Int64("user.id", int64(id))))
	defer span.End()

	user, err := u.userRepository.GetUserById(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "user not found")
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}

	if err := u.userRepository.UpdateUser(ctx, user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update user")
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return u.toUserResp(user), nil
}

func (u *userService) toUserResp(user *domain.User) *dto.UserResponse {
	return &dto.UserResponse{
		Id:        user.Id,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func NewUserService(userRepository repository.UserRepository, tracer trace.Tracer) UserService {
	return &userService{
		userRepository: userRepository,
		tracer:         tracer,
	}
}
