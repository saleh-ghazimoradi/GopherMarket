package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/repository"
	"gorm.io/gorm"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userId uint) (*dto.OrderResponse, error)
	GetOrder(ctx context.Context, userId, orderId uint) (*dto.OrderResponse, error)
	GetOrders(ctx context.Context, userId uint, page, limit int) ([]*dto.OrderResponse, *helper.PaginatedMeta, error)
}

type orderService struct {
	orderRepository    repository.OrderRepository
	cartRepository     repository.CartRepository
	cartItemRepository repository.CartItemRepository
	productRepository  repository.ProductRepository
	db                 *gorm.DB
}

func (o *orderService) CreateOrder(ctx context.Context, userId uint) (*dto.OrderResponse, error) {
	var orderResponse dto.OrderResponse

	err := o.db.Transaction(func(tx *gorm.DB) error {
		cartRepository := o.cartRepository.WithTx(tx)
		cartItemRepository := o.cartItemRepository.WithTx(tx)
		productRepository := o.productRepository.WithTx(tx)
		orderRepository := o.orderRepository.WithTx(tx)

		cart, err := cartRepository.GetCartWithItemsAndProducts(ctx, userId)
		if err != nil {
			return fmt.Errorf("cartRepository.GetCartWithItemsAndProducts: %w", err)
		}

		if len(cart.CartItems) == 0 {
			return errors.New("cart is empty for user")
		}

		var totalAmount float64
		var orderItems []domain.OrderItem

		for i := range cart.CartItems {
			item := &cart.CartItems[i]

			if item.Product.Stock < item.Quantity {
				return fmt.Errorf("not enough stock for product: %s", item.Product.Name)
			}

			item.Product.Stock -= item.Quantity

			if err := productRepository.UpdateProduct(ctx, &item.Product); err != nil {
				return err
			}

			totalAmount += float64(item.Quantity) * item.Product.Price

			orderItems = append(orderItems, domain.OrderItem{
				ProductId: item.Product.Id,
				Quantity:  item.Quantity,
				Price:     item.Product.Price,
			})
		}

		order := &domain.Order{
			UserId:      userId,
			Status:      domain.OrderStatusPending,
			TotalAmount: totalAmount,
			OrderItems:  orderItems,
		}

		if err := orderRepository.CreateOrder(ctx, order); err != nil {
			return fmt.Errorf("orderRepository.CreateOrder: %w", err)
		}

		if err := cartItemRepository.ClearCartItems(ctx, cart.Id); err != nil {
			return fmt.Errorf("cartItemRepository.ClearCartItems: %w", err)
		}

		createdOrder, err := orderRepository.GetOrderById(ctx, order.Id)
		if err != nil {
			return fmt.Errorf("orderRepository.GetOrderById: %w", err)
		}

		resp := o.toOrderResp(createdOrder)
		orderResponse = *resp
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("o.toOrderResp: %w", err)
	}

	return &orderResponse, nil
}

func (o *orderService) GetOrder(ctx context.Context, userId, orderId uint) (*dto.OrderResponse, error) {
	order, err := o.orderRepository.GetOrderByUserId(ctx, userId, orderId)
	if err != nil {
		return nil, fmt.Errorf("o.orderRepository.GetOrderByUserId: %w", err)
	}

	return o.toOrderResp(order), nil
}

func (o *orderService) GetOrders(ctx context.Context, userId uint, page, limit int) ([]*dto.OrderResponse, *helper.PaginatedMeta, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	total, err := o.orderRepository.CountUserOrders(ctx, userId)
	if err != nil {
		return nil, nil, fmt.Errorf("o.orderRepository.CountUserOrders: %w", err)
	}

	orders, err := o.orderRepository.GetUserOrders(ctx, userId, offset, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("o.orderRepository.GetUserOrders: %w", err)
	}

	response := make([]*dto.OrderResponse, len(orders))
	for i := range orders {
		response[i] = o.toOrderResp(orders[i])
	}

	totalPage := int((total + int64(limit) - 1) / int64(limit))

	meta := &helper.PaginatedMeta{
		Page:      page,
		Limit:     limit,
		Total:     total,
		TotalPage: totalPage,
	}

	return response, meta, nil
}

func (o *orderService) toOrderResp(order *domain.Order) *dto.OrderResponse {
	orderItems := make([]dto.OrderItemResponse, len(order.OrderItems))

	for i := range order.OrderItems {
		item := &order.OrderItems[i]
		orderItems[i] = dto.OrderItemResponse{
			Id: item.Id,
			Product: dto.ProductResponse{
				Id:          item.Product.Id,
				CategoryId:  item.Product.CategoryId,
				Name:        item.Product.Name,
				Description: item.Product.Description,
				Price:       item.Product.Price,
				Stock:       item.Product.Stock,
				SKU:         item.Product.SKU,
				IsActive:    item.Product.IsActive,
				Category: dto.CategoryResponse{
					Id:          item.Product.Category.Id,
					Name:        item.Product.Category.Name,
					Description: item.Product.Category.Description,
					IsActive:    item.Product.IsActive,
				},
			},
			Quantity:  item.Quantity,
			Price:     item.Price,
			CreatedAt: item.CreatedAt,
		}
	}
	return &dto.OrderResponse{
		Id:          order.Id,
		UserId:      order.UserId,
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount,
		OrderItems:  orderItems,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}
}

func NewOrderService(orderRepository repository.OrderRepository, cartRepository repository.CartRepository, cartItemRepository repository.CartItemRepository, productRepository repository.ProductRepository, db *gorm.DB) OrderService {
	return &orderService{
		orderRepository:    orderRepository,
		cartRepository:     cartRepository,
		cartItemRepository: cartItemRepository,
		productRepository:  productRepository,
		db:                 db,
	}
}
