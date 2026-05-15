package service

import (
	"context"
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/repository"
)

type UserService interface {
	GetUserById(ctx context.Context, id uint) (*dto.UserResponse, error)
	GetUserProfile(ctx context.Context, id uint) (*dto.UserResponse, error)
	UpdateUserProfile(ctx context.Context, id uint, req *dto.UpdateProfileRequest) (*dto.UserResponse, error)
}

type userService struct {
	userRepository repository.UserRepository
}

func (u *userService) GetUserById(ctx context.Context, id uint) (*dto.UserResponse, error) {
	user, err := u.userRepository.GetUserById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return u.toUserResp(user), nil
}

func (u *userService) GetUserProfile(ctx context.Context, id uint) (*dto.UserResponse, error) {
	user, err := u.userRepository.GetUserById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return u.toUserResp(user), nil
}

func (u *userService) UpdateUserProfile(ctx context.Context, id uint, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, err := u.userRepository.GetUserById(ctx, id)
	if err != nil {
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

func NewUserService(userRepository repository.UserRepository) UserService {
	return &userService{
		userRepository: userRepository,
	}
}
