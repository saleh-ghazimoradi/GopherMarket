package service

import (
	"context"
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/repository"
	"github.com/saleh-ghazimoradi/GopherMarket/utils"
	"time"
)

type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error)
	Logout(ctx context.Context, req *dto.LogoutRequest) error
}

type authService struct {
	userRepository  repository.UserRepository
	cartRepository  repository.CartRepository
	tokenRepository repository.TokenRepository
	cfg             *config.Config
}

func (a *authService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	if _, err := a.userRepository.GetUserByEmail(ctx, req.Email); err == nil {
		return nil, repository.ErrsAlreadyExists
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Role:      domain.Customer,
	}

	if err := a.userRepository.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("faield to create user: %w", err)
	}

	cart := &domain.Cart{UserId: user.Id}
	if err := a.cartRepository.CreateCart(ctx, cart); err != nil {
		return nil, fmt.Errorf("faield to create cart: %w", err)
	}

	return a.generateAuthResponse(ctx, user)
}

func (a *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := a.userRepository.GetUserByEmailAndActive(ctx, req.Email, true)
	if err != nil {
		return nil, fmt.Errorf("faield to get user: %w", err)
	}

	if !utils.CheckPasswordHash(user.Password, req.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	return a.generateAuthResponse(ctx, user)
}

func (a *authService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	claims, err := utils.ValidateToken(req.RefreshToken, a.cfg.JWT.Secret)
	if err != nil {
		return nil, fmt.Errorf("faield to validate refresh token: %w", err)
	}

	refToken, err := a.tokenRepository.GetValidRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid refresh token: %w", err)
	}

	user, err := a.userRepository.GetUserById(ctx, claims.UserId)
	if err != nil {
		return nil, fmt.Errorf("faield to get user: %w", err)
	}

	if err := a.tokenRepository.DeleteTokenById(ctx, refToken.Id); err != nil {
		return nil, fmt.Errorf("faield to delete token: %w", err)
	}

	return a.generateAuthResponse(ctx, user)
}

func (a *authService) Logout(ctx context.Context, req *dto.LogoutRequest) error {
	if err := a.tokenRepository.DeleteToken(ctx, req.RefreshToken); err != nil {
		return fmt.Errorf("failed to log out: %w", err)
	}
	return nil
}

func (a *authService) generateAuthResponse(ctx context.Context, user *domain.User) (*dto.AuthResponse, error) {
	accessToken, refreshToken, err := utils.GenerateToken(a.cfg, user.Id, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshTokenDomain := &domain.RefreshToken{
		UserId:    user.Id,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(a.cfg.JWT.RefreshTokenExpires),
	}

	if err := a.tokenRepository.CreateToken(ctx, refreshTokenDomain); err != nil {
		return nil, fmt.Errorf("faield to create refresh token: %w", err)
	}

	return &dto.AuthResponse{
		User: dto.UserResponse{
			Id:        user.Id,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Phone,
			Role:      string(user.Role),
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func NewAuthService(userRepository repository.UserRepository, cartRepository repository.CartRepository, tokenRepository repository.TokenRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepository:  userRepository,
		cartRepository:  cartRepository,
		tokenRepository: tokenRepository,
		cfg:             cfg,
	}
}
