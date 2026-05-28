package service

import (
	"context"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel/codes"
	"time"

	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/infra/publisher"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/repository"
	"github.com/saleh-ghazimoradi/GopherMarket/pkg/oauth"
	"github.com/saleh-ghazimoradi/GopherMarket/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
)

type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, string, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, string, error)
	GoogleLogin(ctx context.Context, req *dto.GoogleLoginRequest) (*dto.AuthResponse, string, error)
	ChangePassword(ctx context.Context, userId uint, req *dto.ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error
	RefreshToken(ctx context.Context, refreshTokenString string) (*dto.AuthResponse, string, error)
	Logout(ctx context.Context, refreshTokenString string) error
}

type authService struct {
	userRepository       repository.UserRepository
	cartRepository       repository.CartRepository
	tokenRepository      repository.TokenRepository
	resetTokenRepository repository.ResetTokenRepository
	googleOAuth          oauth.Provider
	publisher            publisher.Publisher
	cfg                  *config.Config
	logger               *slog.Logger
	tracer               trace.Tracer
}

func (a *authService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, string, error) {
	ctx, span := a.tracer.Start(ctx, "AuthService.Register",
		trace.WithAttributes(attribute.String("user.email", req.Email)))
	defer span.End()

	if _, err := a.userRepository.GetUserByEmail(ctx, req.Email); err == nil {
		span.SetStatus(codes.Error, "user already exists")
		return nil, "", repository.ErrsAlreadyExists
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to hash password")
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		Email:     req.Email,
		Password:  &hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Role:      domain.Customer,
	}

	if err := a.userRepository.CreateUser(ctx, user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create user")
		return nil, "", fmt.Errorf("faield to create user: %w", err)
	}
	span.SetAttributes(attribute.Int64("user.id", int64(user.Id)))

	cart := &domain.Cart{UserId: user.Id}
	if err := a.cartRepository.CreateCart(ctx, cart); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create cart")
		return nil, "", fmt.Errorf("faield to create cart: %w", err)
	}

	return a.generateAuthResponse(ctx, user)
}

func (a *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, string, error) {
	ctx, span := a.tracer.Start(ctx, "AuthService.Login",
		trace.WithAttributes(attribute.String("user.email", req.Email)))
	defer span.End()

	user, err := a.userRepository.GetUserByEmailAndActive(ctx, req.Email, true)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user")
		return nil, "", fmt.Errorf("faield to get user: %w", err)
	}

	if !utils.CheckPasswordHash(*user.Password, req.Password) {
		span.SetStatus(codes.Error, "invalid credentials")
		return nil, "", fmt.Errorf("invalid credentials")
	}
	span.SetAttributes(attribute.Int64("user.id", int64(user.Id)))

	if err := a.toPublish(ctx, user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to publish user")
		return nil, "", fmt.Errorf("faield to publish user: %w", err)
	}

	return a.generateAuthResponse(ctx, user)
}

func (a *authService) GoogleLogin(ctx context.Context, req *dto.GoogleLoginRequest) (*dto.AuthResponse, string, error) {
	ctx, span := a.tracer.Start(ctx, "AuthService.GoogleLogin")
	defer span.End()

	info, err := a.googleOAuth.Verify(ctx, req.Credential)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "google oauth verify failed")
		return nil, "", fmt.Errorf("google login: %w", err)
	}
	span.SetAttributes(attribute.String("user.email", info.Email))

	user, err := a.userRepository.GetUserByEmail(ctx, info.Email)
	if err != nil && !errors.Is(err, repository.ErrsNotFound) {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to lookup user")
		return nil, "", fmt.Errorf("google login: %w", err)
	}

	provider := "google"
	if user == nil {
		user = &domain.User{
			Email:        info.Email,
			FirstName:    info.GivenName,
			LastName:     info.FamilyName,
			Role:         domain.Customer,
			AuthProvider: &provider,
			IsActive:     true,
		}
		if err := a.userRepository.CreateUser(ctx, user); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to create user")
			return nil, "", fmt.Errorf("google login: create user: %w", err)
		}
		span.SetAttributes(attribute.Int64("user.id", int64(user.Id)))
		cart := &domain.Cart{UserId: user.Id}
		if err := a.cartRepository.CreateCart(ctx, cart); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to create cart")
			return nil, "", fmt.Errorf("google login: create cart: %w", err)
		}
	} else {
		span.SetAttributes(attribute.Int64("user.id", int64(user.Id)))
		if user.AuthProvider == nil {
			user.AuthProvider = &provider
			_ = a.userRepository.UpdateUser(ctx, user)
		}
	}

	return a.generateAuthResponse(ctx, user)
}

func (a *authService) ChangePassword(ctx context.Context, userId uint, req *dto.ChangePasswordRequest) error {
	ctx, span := a.tracer.Start(ctx, "AuthService.ChangePassword",
		trace.WithAttributes(attribute.Int64("user.id", int64(userId))))
	defer span.End()

	user, err := a.userRepository.GetUserById(ctx, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user")
		return fmt.Errorf("faield to get user: %w", err)
	}

	if !utils.CheckPasswordHash(*user.Password, req.OldPassword) {
		span.SetStatus(codes.Error, "invalid credentials")
		return fmt.Errorf("invalid credentials")
	}

	newPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to hash password")
		return fmt.Errorf("faield to hash new password: %w", err)
	}

	user.Password = &newPassword
	if err := a.userRepository.UpdateUser(ctx, user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update user")
		return fmt.Errorf("failed to update user password: %w", err)
	}
	return nil
}

func (a *authService) ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error {
	ctx, span := a.tracer.Start(ctx, "AuthService.ForgotPassword",
		trace.WithAttributes(attribute.String("user.email", req.Email)))
	defer span.End()

	user, err := a.userRepository.GetUserByEmail(ctx, req.Email)
	if err != nil || user == nil || user.Password == nil {
		return nil
	}
	span.SetAttributes(attribute.Int64("user.id", int64(user.Id)))

	code, err := utils.GenerateSecureCode(8)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to generate reset code")
		return fmt.Errorf("failed to generate reset code: %w", err)
	}

	if err := a.resetTokenRepository.Store(ctx, code, user.Id, 15*time.Minute); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to store reset code")
		return fmt.Errorf("failed to store reset code: %w", err)
	}

	resetURL := a.cfg.Application.FrontendURL + "/reset-password"

	eventPayload := &dto.PasswordResetEmailEvent{
		Email:    user.Email,
		ResetURL: resetURL,
		Code:     code,
	}

	if err := a.publisher.Publish(ctx, a.cfg.Event.PasswordResetRequested, eventPayload, map[string]string{}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to publish password reset event")
		return fmt.Errorf("failed to publish password reset event: %w", err)
	}
	return nil
}

func (a *authService) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	ctx, span := a.tracer.Start(ctx, "AuthService.ResetPassword")
	defer span.End()

	userId, err := a.resetTokenRepository.VerifyAndDelete(ctx, req.Code)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid reset token")
		return err
	}
	span.SetAttributes(attribute.Int64("user.id", int64(userId)))

	user, err := a.userRepository.GetUserById(ctx, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "user not found")
		return fmt.Errorf("user not found")
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to hash password")
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.Password = &hashed

	if err := a.userRepository.UpdateUser(ctx, user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update password")
		return fmt.Errorf("faield to update password: %w", err)
	}

	if err := a.tokenRepository.DeleteAllTokensByUserId(ctx, userId); err != nil {
		a.logger.Warn("failed to delete all tokens", "error", err)
	}

	return nil
}

func (a *authService) RefreshToken(ctx context.Context, refreshTokenString string) (*dto.AuthResponse, string, error) {
	ctx, span := a.tracer.Start(ctx, "AuthService.RefreshToken")
	defer span.End()

	claims, err := utils.ValidateToken(refreshTokenString, a.cfg.JWT.Secret)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid refresh token")
		return nil, "", fmt.Errorf("faield to validate refresh token: %w", err)
	}
	span.SetAttributes(attribute.Int64("user.id", int64(claims.UserId)))

	refToken, err := a.tokenRepository.GetValidRefreshToken(ctx, refreshTokenString)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid or expired refresh token")
		return nil, "", fmt.Errorf("failed to get valid refresh token: %w", err)
	}

	user, err := a.userRepository.GetUserById(ctx, claims.UserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "user not found")
		return nil, "", fmt.Errorf("faield to get user: %w", err)
	}

	if err := a.tokenRepository.DeleteTokenById(ctx, refToken.Id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to revoke old token")
		return nil, "", fmt.Errorf("faield to delete token: %w", err)
	}

	return a.generateAuthResponse(ctx, user)
}

func (a *authService) Logout(ctx context.Context, refreshTokenString string) error {
	ctx, span := a.tracer.Start(ctx, "AuthService.Logout")
	defer span.End()

	if err := a.tokenRepository.DeleteToken(ctx, refreshTokenString); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to logout")
		return fmt.Errorf("failed to log out: %w", err)
	}
	return nil
}

func (a *authService) toPublish(ctx context.Context, user *domain.User) error {
	if err := a.publisher.Publish(ctx, a.cfg.Event.UserLoggedIn, user, map[string]string{}); err != nil {
		return fmt.Errorf("faield to publish refresh token: %w", err)
	}
	return nil
}

func (a *authService) generateAuthResponse(ctx context.Context, user *domain.User) (*dto.AuthResponse, string, error) {
	accessToken, refreshTokenStr, err := utils.GenerateToken(a.cfg, user.Id, user.Email, string(user.Role))
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	refreshTokenDomain := &domain.RefreshToken{
		UserId:    user.Id,
		Token:     refreshTokenStr,
		ExpiresAt: time.Now().Add(a.cfg.JWT.RefreshTokenExpires),
	}

	if err := a.tokenRepository.CreateToken(ctx, refreshTokenDomain); err != nil {
		return nil, "", fmt.Errorf("faield to create refresh token: %w", err)
	}

	authResp := &dto.AuthResponse{
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
		AccessToken: accessToken,
	}
	return authResp, refreshTokenStr, nil
}

func NewAuthService(
	userRepository repository.UserRepository,
	cartRepository repository.CartRepository,
	tokenRepository repository.TokenRepository,
	resetTokenRepository repository.ResetTokenRepository,
	googleOAuth oauth.Provider,
	publisher publisher.Publisher,
	cfg *config.Config,
	logger *slog.Logger,
	tracer trace.Tracer,
) AuthService {
	return &authService{
		userRepository:       userRepository,
		cartRepository:       cartRepository,
		tokenRepository:      tokenRepository,
		resetTokenRepository: resetTokenRepository,
		googleOAuth:          googleOAuth,
		publisher:            publisher,
		cfg:                  cfg,
		logger:               logger,
		tracer:               tracer,
	}
}
