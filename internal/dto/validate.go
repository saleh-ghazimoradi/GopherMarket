package dto

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"time"
)

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

func validateProductId(v *helper.Validator, productId uint) {
	v.Check(productId > 0, "product_id", "Product ID must be provided")
}

func validateDiscountType(v *helper.Validator, discountType domain.DiscountType) {
	isValidType := discountType == domain.DiscountPercentage || discountType == domain.DiscountFixed
	v.Check(isValidType, "discount_type", "Discount type must be percentage, fixed")
}

func validateDiscountValue(v *helper.Validator, discountType domain.DiscountType, value float64) {
	v.Check(value > 0, "discount_value", "Discount value must be greater than 0")
	if discountType == domain.DiscountPercentage {
		v.Check(value <= 100, "discount_value", "Percentage discount cannot exceed 100%")
	}
}

func validateDiscountTimeline(v *helper.Validator, startTime, endTime time.Time) {
	v.Check(!startTime.IsZero(), "start_time", "Start time must be provided")
	v.Check(!endTime.IsZero(), "end_time", "End time must be provided")
	v.Check(endTime.After(startTime), "end_time", "End time must be after the start time")
	v.Check(endTime.After(time.Now()), "end_time", "End time cannot be in the past")
}

func validateCategoryName(v *helper.Validator, name string) {
	v.Check(helper.NotBlank(name), "name", "Category name must be provided")
}

func validateCategoryId(v *helper.Validator, id uint) {
	v.Check(id > 0, "id", "Category ID must be greater than 0")
}

func validateProductName(v *helper.Validator, name string) {
	v.Check(helper.NotBlank(name), "name", "Product name must be provided")
}

func validateProductPrice(v *helper.Validator, price float64) {
	v.Check(price > 0, "price", "Price must be greater than 0")
}

func validateStock(v *helper.Validator, stock int) {
	v.Check(stock > 0, "stock", "Stock must be greater than 0")
}

func validateProductSKU(v *helper.Validator, sku string) {
	v.Check(helper.NotBlank(sku), "sku", "Product sku must be provided")
}

func validateQuery(v *helper.Validator, query string) {
	v.Check(helper.MinChars(query, 1), "query", "Query must be at least 1 characters")
}

func validateToken(v *helper.Validator, token string) {
	v.Check(helper.NotBlank(token), "token", "Token must be provided")
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

func ValidateUpdateProfileRequest(v *helper.Validator, req *UpdateProfileRequest) {
	if req.FirstName != nil {
		validateFirstNameAndLastName(v, *req.FirstName)
	}
	if req.LastName != nil {
		validateFirstNameAndLastName(v, *req.LastName)
	}
}

func ValidateCreateCategoryRequest(v *helper.Validator, req *CreateCategoryRequest) {
	validateCategoryName(v, req.Name)
}

func ValidateUpdateCategoryRequest(v *helper.Validator, req *UpdateCategoryRequest) {
	if req.Name != nil {
		validateCategoryName(v, *req.Name)
	}
}

func ValidateCreateProductRequest(v *helper.Validator, req *CreateProductRequest) {
	validateCategoryId(v, req.CategoryId)
	validateProductName(v, req.Name)
	validateProductPrice(v, req.Price)
	validateStock(v, req.Stock)
	validateProductSKU(v, req.SKU)
}

func ValidateUpdateProductRequest(v *helper.Validator, req *UpdateProductRequest) {
	if req.Name != nil {
		validateProductName(v, *req.Name)
	}
	if req.CategoryId != nil {
		validateCategoryId(v, *req.CategoryId)
	}
	if req.Price != nil {
		validateProductPrice(v, *req.Price)
	}
	if req.Stock != nil {
		validateStock(v, *req.Stock)
	}
}

func ValidateQuery(v *helper.Validator, req *SearchProductsRequest) {
	if req.Query != "" {
		validateQuery(v, req.Query)
	}
}

func ValidateGoogleLoginRequest(v *helper.Validator, req *GoogleLoginRequest) {
	v.Check(req.Credential != "", "credential", "must be provided")
}

func ValidateForgotPasswordRequest(v *helper.Validator, req *ForgotPasswordRequest) {
	validateEmail(v, req.Email)
}

func ValidateResetPasswordRequest(v *helper.Validator, req *ResetPasswordRequest) {
	validateToken(v, req.Code)
	validatePassword(v, req.Password)
}

func ValidateChangePasswordRequest(v *helper.Validator, req *ChangePasswordRequest) {
	validatePassword(v, req.NewPassword)
}

func ValidateCreateDiscountRequest(v *helper.Validator, req *CreateDiscountRequest) {
	validateProductId(v, req.ProductId)
	validateDiscountType(v, req.DiscountType)
	validateDiscountValue(v, req.DiscountType, req.DiscountValue)
	validateDiscountTimeline(v, req.StartTime, req.EndTime)
}
