package dto

import "github.com/saleh-ghazimoradi/GopherMarket/internal/helper"

func validateFirstNameAndLastName(v *helper.Validator, name string) {
	v.Check(helper.NotBlank(name), "name", "First and last names must be provided")
	v.Check(helper.MaxChars(name, 100), "name", "First and last names must be less than 100 characters")
}

func validateEmail(v *helper.Validator, email string) {
	v.Check(helper.NotBlank(email), "email", "Email must be provided")
	v.Check(helper.Matches(email, helper.EmailRX), "email", "Email must be valid")
	v.Check(helper.MaxChars(email, 100), "email", "Email must be less than 100 characters")
}

func validatePassword(v *helper.Validator, password string) {
	v.Check(helper.NotBlank(password), "password", "Password must be provided")
	v.Check(helper.MinChars(password, 8), "password", "Password must be at least 8 characters")
	v.Check(helper.MaxChars(password, 72), "password", "Password must be less than 72 characters")
}

func validateRefreshToken(v *helper.Validator, refreshToken string) {
	v.Check(helper.NotBlank(refreshToken), "refresh_token", "Refresh token must be provided")
}

func ValidateRegisterRequest(v *helper.Validator, req *RegisterRequest) {
	validateFirstNameAndLastName(v, req.FirstName)
	validateFirstNameAndLastName(v, req.LastName)
	validateEmail(v, req.Email)
	validatePassword(v, req.Password)
}

func ValidateLoginRequest(v *helper.Validator, req *LoginRequest) {
	validateEmail(v, req.Email)
	validatePassword(v, req.Password)
}

func ValidateRefreshTokenRequest(v *helper.Validator, req *RefreshTokenRequest) {
	validateRefreshToken(v, req.RefreshToken)
}

func ValidateLogoutRequest(v *helper.Validator, req *LogoutRequest) {
	validateRefreshToken(v, req.RefreshToken)
}

func ValidateProfileRequest(v *helper.Validator, req *UpdateProfileRequest) {
	validateFirstNameAndLastName(v, *req.FirstName)
	validateFirstNameAndLastName(v, *req.LastName)
}
