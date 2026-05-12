package handler

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"github.com/saleh-ghazimoradi/GopherMarket/utils"
	"net/http"
)

type CartHandler struct {
	cartService service.CartService
}

func (c *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userId, exists := utils.UserIdFromContext(r.Context())
	if !exists {
		helper.UnauthorizedResponse(w, "unauthorized")
		return
	}

	cart, err := c.cartService.GetCart(r.Context(), userId)
	if err != nil {
		helper.NotFoundResponse(w, "Cart not found")
		return
	}

	helper.SuccessResponse(w, "Cart successfully retrieved", cart)
}

func (c *CartHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	userId, exists := utils.UserIdFromContext(r.Context())
	if !exists {
		helper.UnauthorizedResponse(w, "unauthorized")
		return
	}

	var payload dto.AddToCartRequest
	if err := helper.ReadJSON(w, r, &payload); err != nil {
		helper.BadRequestResponse(w, "invalid given payload", err)
		return
	}

	cart, err := c.cartService.AddToCart(r.Context(), userId, &payload)
	if err != nil {
		helper.InternalServerError(w, "failed to add item to cart", err)
		return
	}

	helper.SuccessResponse(w, "Item successfully added to cart", cart)
}

func (c *CartHandler) UpdateCart(w http.ResponseWriter, r *http.Request) {
	userId, exists := utils.UserIdFromContext(r.Context())
	if !exists {
		helper.UnauthorizedResponse(w, "unauthorized")
		return
	}

	id, err := helper.ReadParams(r)
	if err != nil {
		helper.BadRequestResponse(w, "invalid cart id", err)
		return
	}

	var payload dto.UpdateCartItemRequest
	if err := helper.ReadJSON(w, r, &payload); err != nil {
		helper.BadRequestResponse(w, "invalid given payload", err)
		return
	}

	cart, err := c.cartService.UpdateCartItem(r.Context(), userId, id, &payload)
	if err != nil {
		helper.InternalServerError(w, "failed to update item to cart", err)
		return
	}

	helper.SuccessResponse(w, "Cart successfully updated to cart", cart)
}

func (c *CartHandler) RemoveCart(w http.ResponseWriter, r *http.Request) {
	userId, exists := utils.UserIdFromContext(r.Context())
	if !exists {
		helper.UnauthorizedResponse(w, "unauthorized")
		return
	}

	id, err := helper.ReadParams(r)
	if err != nil {
		helper.BadRequestResponse(w, "invalid cart item id", err)
		return
	}

	if err := c.cartService.RemoveFromCart(r.Context(), userId, id); err != nil {
		helper.InternalServerError(w, "failed to remove item from cart", err)
		return
	}

	helper.SuccessResponse(w, "Cart successfully removed from cart", nil)
}

func NewCartHandler(cartService service.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}
