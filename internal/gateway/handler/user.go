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
