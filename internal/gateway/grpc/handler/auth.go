package handler

import (
	"context"
	"errors"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/grpc/protos"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/repository"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthGrpcHandler struct {
	protos.UnimplementedAuthServiceServer
	authService service.AuthService
}

func (a *AuthGrpcHandler) Register(ctx context.Context, req *protos.RegisterRequest) (*protos.AuthResponse, error) {
	registerDTO := &dto.RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	}

	resp, refreshToken, err := a.authService.Register(ctx, registerDTO)
	if err != nil {
		if errors.Is(err, repository.ErrsAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "user already exists: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error registering a user: %v", err)
	}

	return a.mapToAuthResponse(resp, refreshToken), nil
}

func (a *AuthGrpcHandler) Login(ctx context.Context, req *protos.LoginRequest) (*protos.AuthResponse, error) {
	loginDTO := &dto.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, refreshToken, err := a.authService.Login(ctx, loginDTO)
	if err != nil {
		if errors.Is(err, repository.ErrsNotFound) {
			return nil, status.Errorf(codes.NotFound, "invalid email or password")
		}
		return nil, status.Errorf(codes.Internal, "error logging in a user: %v", err)
	}
	return a.mapToAuthResponse(resp, refreshToken), nil
}

func (a *AuthGrpcHandler) GoogleLogin(ctx context.Context, req *protos.GoogleLoginRequest) (*protos.AuthResponse, error) {
	googleLoginDTO := &dto.GoogleLoginRequest{
		Credential: req.Credential,
	}

	resp, refreshToken, err := a.authService.GoogleLogin(ctx, googleLoginDTO)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error logging in through OAuth: %v", err)
	}
	return a.mapToAuthResponse(resp, refreshToken), nil
}

func (a *AuthGrpcHandler) ChangePassword(ctx context.Context, req *protos.ChangePasswordRequest) (*emptypb.Empty, error) {
	changePasswordDTO := &dto.ChangePasswordRequest{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}
	if err := a.authService.ChangePassword(ctx, uint(req.UserId), changePasswordDTO); err != nil {
		if errors.Is(err, repository.ErrsNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error changing password: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (a *AuthGrpcHandler) ForgotPassword(ctx context.Context, req *protos.ForgotPasswordRequest) (*emptypb.Empty, error) {
	forgotPasswordDTO := &dto.ForgotPasswordRequest{
		Email: req.Email,
	}

	if err := a.authService.ForgotPassword(ctx, forgotPasswordDTO); err != nil {
		return nil, status.Errorf(codes.Internal, "error executing forgot password sequence: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (a *AuthGrpcHandler) ResetPassword(ctx context.Context, req *protos.ResetPasswordRequest) (*emptypb.Empty, error) {
	resetPasswordDTO := &dto.ResetPasswordRequest{
		Code:     req.Code,
		Password: req.Password,
	}

	if err := a.authService.ResetPassword(ctx, resetPasswordDTO); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired validation token: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (a *AuthGrpcHandler) RefreshToken(ctx context.Context, req *protos.RefreshTokenRequest) (*protos.AuthResponse, error) {
	resp, newRefreshToken, err := a.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid refresh session context: %v", err)
	}

	return a.mapToAuthResponse(resp, newRefreshToken), nil
}

func (a *AuthGrpcHandler) Logout(ctx context.Context, req *protos.LogoutRequest) (*emptypb.Empty, error) {
	if err := a.authService.Logout(ctx, req.RefreshToken); err != nil {
		return nil, status.Errorf(codes.Internal, "error logging out user: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (a *AuthGrpcHandler) mapToAuthResponse(auth *dto.AuthResponse, refreshToken string) *protos.AuthResponse {
	return &protos.AuthResponse{
		User: &protos.UserResponse{
			Id:        uint64(auth.User.Id),
			Email:     auth.User.Email,
			FirstName: auth.User.FirstName,
			LastName:  auth.User.LastName,
			Phone:     auth.User.Phone,
			Role:      auth.User.Role,
			IsActive:  auth.User.IsActive,
			CreatedAt: timestamppb.New(auth.User.CreatedAt),
			UpdatedAt: timestamppb.New(auth.User.UpdatedAt),
		},
		AccessToken:  auth.AccessToken,
		RefreshToken: refreshToken,
	}
}

func NewAuthGrpcHandler(authService service.AuthService) *AuthGrpcHandler {
	return &AuthGrpcHandler{
		authService: authService,
	}
}
