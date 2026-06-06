package handler

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"net/http"
)

type DiscountHandler struct {
	discountService service.DiscountService
}

func (d *DiscountHandler) CreateDiscount(w http.ResponseWriter, r *http.Request) {
	payload := dto.CreateDiscountRequest{}
	if err := helper.ReadJSON(w, r, &payload); err != nil {
		helper.BadRequestResponse(w, "Invalid given payload", err)
		return
	}

	v := helper.NewValidator()
	dto.ValidateCreateDiscountRequest(v, &payload)
	if !v.Valid() {
		helper.FailedValidationResponse(w, "Payload is not valid")
		return
	}

	discount, err := d.discountService.CreateDiscount(r.Context(), &payload)
	if err != nil {
		helper.InternalServerError(w, "Error creating discount", err)
		return
	}

	helper.SuccessResponse(w, "Discount successfully created", discount)
}

func (d *DiscountHandler) DeleteDiscount(w http.ResponseWriter, r *http.Request) {
	id, err := helper.ReadParams(r, "id")
	if err != nil {
		helper.NotFoundResponse(w, "record not found")
		return
	}

	productId, err := helper.ReadParams(r, "productId")
	if err != nil {
		helper.NotFoundResponse(w, "record not found")
		return
	}

	if err := d.discountService.DeleteDiscount(r.Context(), id, productId); err != nil {
		helper.InternalServerError(w, "Error deleting discount", err)
		return
	}

	helper.SuccessResponse(w, "Discount successfully deleted", nil)
}

func NewDiscountHandler(discountService service.DiscountService) *DiscountHandler {
	return &DiscountHandler{
		discountService: discountService,
	}
}
