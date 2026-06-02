package handler

import (
	"context"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"

	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/grpc/protos"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserGrpcHandler struct {
	protos.UnimplementedUserServiceServer
	userService service.UserService
}

func (u *UserGrpcHandler) GetUserById(ctx context.Context, req *protos.GetUserRequest) (*protos.GetUserResponse, error) {
	user, err := u.userService.GetUserById(ctx, uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}
	return u.mapToResponse(user), nil
}

func (u *UserGrpcHandler) GetUserProfile(ctx context.Context, req *protos.GetUserProfileRequest) (*protos.GetUserResponse, error) {
	user, err := u.userService.GetUserById(ctx, uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "profile not found: %v", err)
	}

	return u.mapToResponse(user), nil
}

func (u *UserGrpcHandler) UpdateUserProfile(ctx context.Context, req *protos.UpdateUserProfileRequest) (*protos.GetUserResponse, error) {
	updateDTO := &dto.UpdateProfileRequest{}

	if req.FirstName != nil {
		updateDTO.FirstName = req.FirstName
	}

	if req.LastName != nil {
		updateDTO.LastName = req.LastName
	}

	if req.Phone != nil {
		updateDTO.Phone = req.Phone
	}

	updatedProfile, err := u.userService.UpdateUserProfile(ctx, uint(req.Id), updateDTO)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user profile: %v", err)
	}

	return u.mapToResponse(updatedProfile), nil
}

func (u *UserGrpcHandler) mapToResponse(user *dto.UserResponse) *protos.GetUserResponse {
	return &protos.GetUserResponse{
		Id:        uint64(user.Id),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}

func NewUserGrpcHandler(userService service.UserService) *UserGrpcHandler {
	return &UserGrpcHandler{
		userService: userService,
	}
}
