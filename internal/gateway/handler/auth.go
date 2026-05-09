package handler

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"net/http"
)

type AuthHandler struct {
	authService service.AuthService
}

func (a *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var payload dto.RegisterRequest
	if err := helper.ReadJSON(w, r, &payload); err != nil {
		helper.BadRequestResponse(w, "Invalid given payload", err)
		return
	}

	v := helper.NewValidator()
	dto.ValidateRegisterRequest(v, &payload)
	if !v.Valid() {
		helper.FailedValidationResponse(w, "payload is not valid")
		return
	}

	user, err := a.authService.Register(r.Context(), &payload)
	if err != nil {
		helper.InternalServerError(w, "failed to register user", err)
		return
	}

	helper.CreatedResponse(w, "user successfully registered", user)
}

func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var payload dto.LoginRequest
	if err := helper.ReadJSON(w, r, &payload); err != nil {
		helper.BadRequestResponse(w, "Invalid given payload", err)
		return
	}

	v := helper.NewValidator()
	dto.ValidateLoginRequest(v, &payload)
	if !v.Valid() {
		helper.FailedValidationResponse(w, "payload is not valid")
		return
	}

	user, err := a.authService.Login(r.Context(), &payload)
	if err != nil {
		helper.InternalServerError(w, "failed to login", err)
		return
	}

	helper.SuccessResponse(w, "user successfully login", user)
}

func (a *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var payload dto.RefreshTokenRequest
	if err := helper.ReadJSON(w, r, &payload); err != nil {
		helper.BadRequestResponse(w, "Invalid given payload", err)
		return
	}

	v := helper.NewValidator()
	dto.ValidateRefreshTokenRequest(v, &payload)
	if !v.Valid() {
		helper.FailedValidationResponse(w, "payload is not valid")
		return
	}

	refreshToken, err := a.authService.RefreshToken(r.Context(), &payload)
	if err != nil {
		helper.InternalServerError(w, "failed to refresh token", err)
		return
	}

	helper.SuccessResponse(w, "refresh token successfully", refreshToken)
}

func (a *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var payload dto.LogoutRequest
	if err := helper.ReadJSON(w, r, &payload); err != nil {
		helper.BadRequestResponse(w, "Invalid given payload", err)
		return
	}

	v := helper.NewValidator()
	dto.ValidateLogoutRequest(v, &payload)
	if !v.Valid() {
		helper.FailedValidationResponse(w, "payload is not valid")
		return
	}

	if err := a.authService.Logout(r.Context(), &payload); err != nil {
		helper.InternalServerError(w, "failed to logout", err)
	}

	helper.SuccessResponse(w, "user successfully logged out", nil)
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}
