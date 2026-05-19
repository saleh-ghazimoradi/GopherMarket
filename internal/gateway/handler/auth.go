package handler

import (
	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"net/http"
)

type AuthHandler struct {
	authService service.AuthService
	cfg         *config.Config
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account with email and password
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "User registration data"
// @Success      201 {object} helper.Response{data=dto.AuthResponse}
// @Failure      400 {object} helper.Response "Invalid request data or user already exists"
// @Router       /auth/register [post]
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

	authResp, refreshToken, err := a.authService.Register(r.Context(), &payload)
	if err != nil {
		helper.InternalServerError(w, "failed to register user", err)
		return
	}

	a.setRefreshTokenCookie(w, refreshToken)
	helper.CreatedResponse(w, "user successfully registered", authResp)
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user with email and password
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "User login credentials"
// @Success      200 {object} helper.Response{data=dto.AuthResponse} "Login successfully"
// @Failure      401 {object} helper.Response "Invalid credentials"
// @Router       /auth/login [post]
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

	authResp, refreshToken, err := a.authService.Login(r.Context(), &payload)
	if err != nil {
		helper.InternalServerError(w, "failed to login", err)
		return
	}

	a.setRefreshTokenCookie(w, refreshToken)
	helper.SuccessResponse(w, "user successfully login", authResp)
}

// GoogleLogin godoc
// @Summary      Google OAuth login
// @Description  Authenticate user with a Google ID token. Creates a new account if the email is not registered.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body dto.GoogleLoginRequest true "Google ID token"
// @Success      200 {object} helper.Response{data=dto.AuthResponse} "Login successfully"
// @Failure      400 {object} helper.Response "Invalid request data or invalid Google token"
// @Failure      500 {object} helper.Response "Internal server error"
// @Router       /v1/auth/google [post]
func (a *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var payload dto.GoogleLoginRequest
	if err := helper.ReadJSON(w, r, &payload); err != nil {
		helper.BadRequestResponse(w, "Invalid given payload", err)
		return
	}
	v := helper.NewValidator()
	dto.ValidateGoogleLoginRequest(v, &payload)
	if !v.Valid() {
		helper.FailedValidationResponse(w, "payload is not valid")
		return
	}

	authResp, refreshToken, err := a.authService.GoogleLogin(r.Context(), &payload)
	if err != nil {
		helper.InternalServerError(w, "failed to login", err)
		return
	}

	a.setRefreshTokenCookie(w, refreshToken)
	helper.SuccessResponse(w, "Login successfully", authResp)
}

// ForgotPassword godoc
// @Summary      Request password reset email
// @Description  Sends a password reset link to the email if it exists and belongs to a password‑based account. The response always returns success to prevent user enumeration.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body dto.ForgotPasswordRequest true "User email"
// @Success      200 {object} helper.Response "If the email exists, a reset link has been sent"
// @Failure      400 {object} helper.Response "Invalid request data"
// @Failure      500 {object} helper.Response "Internal server error"
// @Router       /v1/auth/forgot-password [post]
func (a *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var payload dto.ForgotPasswordRequest
	if err := helper.ReadJSON(w, r, &payload); err != nil {
		helper.BadRequestResponse(w, "Invalid given payload", err)
		return
	}

	v := helper.NewValidator()
	dto.ValidateForgotPasswordRequest(v, &payload)
	if !v.Valid() {
		helper.FailedValidationResponse(w, "payload is not valid")
		return
	}

	if err := a.authService.ForgotPassword(r.Context(), &payload); err != nil {
		helper.InternalServerError(w, "could not process request", err)
		return
	}

	helper.SuccessResponse(w, "A reset password link has been sent", nil)
}

// ResetPassword godoc
// @Summary      Reset password with token
// @Description  Sets a new password using the token received by email. All existing refresh tokens are invalidated.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body dto.ResetPasswordRequest true "Reset token and new password"
// @Success      200 {object} helper.Response "Password successfully reset"
// @Failure      400 {object} helper.Response "Invalid request data or invalid/expired token"
// @Router       /v1/auth/reset-password [post]
func (a *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var payload dto.ResetPasswordRequest
	if err := helper.ReadJSON(w, r, &payload); err != nil {
		helper.BadRequestResponse(w, "Invalid given payload", err)
		return
	}

	v := helper.NewValidator()
	dto.ValidateResetPasswordRequest(v, &payload)
	if !v.Valid() {
		helper.FailedValidationResponse(w, "payload is not valid")
		return
	}

	if err := a.authService.ResetPassword(r.Context(), &payload); err != nil {
		helper.BadRequestResponse(w, "password reset failed", err)
		return
	}

	helper.SuccessResponse(w, "password successfully reset", nil)
}

// RefreshToken docs
// @Summary Refresh access token
// @Description Get a new access token using refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} helper.Response{data=dto.AuthResponse} "Token refreshed successfully"
// @Failure 401 {object} helper.Response "Invalid refresh token"
// @Router /auth/refresh [post]
func (a *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		helper.UnauthorizedResponse(w, "refresh token missing")
		return
	}

	tokenStr := cookie.Value

	authResp, newRefreshToken, err := a.authService.RefreshToken(r.Context(), tokenStr)
	if err != nil {
		helper.InternalServerError(w, "failed to refresh token", err)
		return
	}

	a.setRefreshTokenCookie(w, newRefreshToken)
	helper.SuccessResponse(w, "refresh token successfully", authResp)
}

// Logout docs
// @Summary User logout
// @Description Invalidate refresh token and logout user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token to invalidate"
// @Success 200 {object} helper.Response "Logout successful"
// @Failure 400 {object} helper.Response "Invalid request data"
// @Router /auth/logout [post]
func (a *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		// No cookie to clear, but still send a delete cookie just in case.
		a.clearRefreshTokenCookie(w)
		helper.SuccessResponse(w, "user successfully logged out", nil)
		return
	}

	if err := a.authService.Logout(r.Context(), cookie.Value); err != nil {
		// Even if DB deletion fails, clear the cookie.
		a.clearRefreshTokenCookie(w)
		helper.InternalServerError(w, "failed to logout", err)
		return
	}

	a.clearRefreshTokenCookie(w)
	helper.SuccessResponse(w, "user successfully logged out", nil)
}

func (a *AuthHandler) setRefreshTokenCookie(w http.ResponseWriter, token string) {
	//TODO: Fix the issue of Secure field. It must be set to true in production. Remove the secure check when in production and switch to HTTPS!!!!
	secure := a.cfg.Application.Environment != "development"
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/v1/auth",
		Domain:   "",
		MaxAge:   int(a.cfg.JWT.RefreshTokenExpires.Seconds()),
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (a *AuthHandler) clearRefreshTokenCookie(w http.ResponseWriter) {
	//TODO: Fix the issue of Secure field. It must be set to true in production. Remove the secure check when in production and switch to HTTPS!!!!
	secure := a.cfg.Application.Environment != "development"
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/v1/auth",
		MaxAge:   -1,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func NewAuthHandler(authService service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cfg:         cfg,
	}
}
