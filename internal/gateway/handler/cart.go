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

// GetCart docs
// @Summary Get user's cart
// @Description Retrieve current user's shopping cart with all items
// @Tags Carts
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helper.Response{data=dto.CartResponse} "Cart retrieved successfully"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Failure 404 {object} helper.Response "Cart not found"
// @Router /carts [get]
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

// AddToCart docs
// @Summary Add item to cart
// @Description Add a product to the user's shopping cart
// @Tags Carts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.AddToCartRequest true "Item to add to cart"
// @Success 200 {object} helper.Response{data=dto.CartResponse} "Item added to cart successfully"
// @Failure 400 {object} helper.Response "Invalid request data or insufficient stock"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Router /carts/items [post]
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

// UpdateCart docs
// @Summary Update cart item quantity
// @Description Update the quantity of an item in the user's cart
// @Tags Carts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Cart Item ID"
// @Param request body dto.UpdateCartItemRequest true "New quantity"
// @Success 200 {object} helper.Response{data=dto.CartResponse} "Cart item updated successfully"
// @Failure 400 {object} helper.Response "Invalid request data or insufficient stock"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Router /carts/items/{id} [put]
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

// RemoveCart docs
// @Summary Remove item from cart
// @Description Remove an item from the user's shopping cart
// @Tags Carts
// @Security BearerAuth
// @Param id path int true "Cart Item ID"
// @Success 200 {object} helper.Response "Item removed from cart successfully"
// @Failure 400 {object} helper.Response "Invalid cart item ID"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Router /carts/items/{id} [delete]
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
