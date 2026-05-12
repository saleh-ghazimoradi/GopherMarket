package handler

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"github.com/saleh-ghazimoradi/GopherMarket/utils"
	"net/http"
	"strconv"
)

type OrderHandler struct {
	orderService service.OrderService
}

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

func (o *OrderHandler) GetUserOrder(w http.ResponseWriter, r *http.Request) {
	userId, exist := utils.UserIdFromContext(r.Context())
	if !exist {
		helper.UnauthorizedResponse(w, "Unauthorized")
		return
	}

	id, err := helper.ReadParams(r)
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

func (o *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userId, exist := utils.UserIdFromContext(r.Context())
	if !exist {
		helper.UnauthorizedResponse(w, "Unauthorized")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

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
