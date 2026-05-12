package handler

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"github.com/saleh-ghazimoradi/GopherMarket/utils"
	"net/http"
)

type UserHandler struct {
	userService service.UserService
}

// GetUserProfile docs
// @Summary Get user profile
// @Description Get current authenticated user's profile information
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helper.Response{data=dto.UserResponse} "Profile retrieved successfully"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Failure 404 {object} helper.Response "User not found"
// @Router /users/profile [get]
func (u *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	id, exists := utils.UserIdFromContext(r.Context())
	if !exists {
		helper.UnauthorizedResponse(w, "Unauthorized")
		return
	}

	profile, err := u.userService.GetUserProfile(r.Context(), id)
	if err != nil {
		helper.InternalServerError(w, "failed to get profile", err)
		return
	}

	helper.SuccessResponse(w, "profile successfully retrieved", profile)
}

// UpdateUserProfile docs
// @Summary Update user profile
// @Description Update current authenticated user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateProfileRequest true "Profile update data"
// @Success 200 {object} helper.Response{data=dto.UserResponse} "Profile updated successfully"
// @Failure 400 {object} helper.Response "Invalid request data"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Router /users/profile [put]
func (u *UserHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	id, exists := utils.UserIdFromContext(r.Context())
	if !exists {
		helper.UnauthorizedResponse(w, "Unauthorized")
		return
	}

	var payload dto.UpdateProfileRequest
	if err := helper.ReadJSON(w, r, &payload); err != nil {
		helper.BadRequestResponse(w, "invalid given payload", err)
		return
	}

	v := helper.NewValidator()
	dto.ValidateUpdateProfileRequest(v, &payload)
	if !v.Valid() {
		helper.FailedValidationResponse(w, "payload is invalid")
		return
	}

	profile, err := u.userService.UpdateUserProfile(r.Context(), id, &payload)
	if err != nil {
		helper.InternalServerError(w, "failed to update profile", err)
		return
	}

	helper.SuccessResponse(w, "profile successfully updated", profile)
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}
