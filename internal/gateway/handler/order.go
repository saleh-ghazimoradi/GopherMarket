package handler

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"github.com/saleh-ghazimoradi/GopherMarket/utils"
	"net/http"
)

type OrderHandler struct {
	orderService service.OrderService
}

// CreateOrder dcs
// @Summary Create an order
// @Description Create an order from the current user's cart
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Success 201 {object} helper.Response{data=dto.OrderResponse} "Order created successfully"
// @Failure 400 {object} helper.Response "Cart is empty or insufficient stock"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Router /orders [post]
func (o *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userId, exist := utils.UserIdFromContext(r.Context())
	if !exist {
		helper.UnauthorizedResponse(w, "Unauthorized")
		return
	}

	order, err := o.orderService.CreateOrder(r.Context(), userId)
	if err != nil {
		helper.InternalServerError(w, "failed to create order", err)
		return
	}

	helper.CreatedResponse(w, "Order successfully created", order)
}

// GetUserOrder docs
// @Summary Get order by ID
// @Description Retrieve detailed information about a specific order
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} helper.Response{data=dto.OrderResponse} "Order retrieved successfully"
// @Failure 400 {object} helper.Response "Invalid order ID"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Failure 404 {object} helper.Response "Order not found"
// @Router /orders/{id} [get]
func (o *OrderHandler) GetUserOrder(w http.ResponseWriter, r *http.Request) {
	userId, exist := utils.UserIdFromContext(r.Context())
	if !exist {
		helper.UnauthorizedResponse(w, "Unauthorized")
		return
	}

	id, err := helper.ReadParams(r, "id")
	if err != nil {
		helper.BadRequestResponse(w, "Invalid order id", err)
		return
	}

	order, err := o.orderService.GetOrder(r.Context(), userId, id)
	if err != nil {
		helper.NotFoundResponse(w, "order not found")
		return
	}

	helper.SuccessResponse(w, "Order successfully retrieved", order)
}

// GetUserOrders docs
// @Summary Get user's orders
// @Description Retrieve paginated list of user's orders
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} helper.PaginatedResponse{data=[]dto.OrderResponse} "Orders retrieved successfully"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /orders [get]
func (o *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userId, exist := utils.UserIdFromContext(r.Context())
	if !exist {
		helper.UnauthorizedResponse(w, "Unauthorized")
		return
	}

	page, _ := helper.ReadQueryParam(r, "page")
	limit, _ := helper.ReadQueryParam(r, "limit")

	orders, meta, err := o.orderService.GetOrders(r.Context(), userId, page, limit)
	if err != nil {
		helper.InternalServerError(w, "failed to get orders", err)
		return
	}

	helper.PaginatedSuccessResponse(w, "Orders successfully retrieved", orders, *meta)
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}
