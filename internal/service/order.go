package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	tracer             trace.Tracer
}

func (o *orderService) CreateOrder(ctx context.Context, userId uint) (*dto.OrderResponse, error) {
	ctx, span := o.tracer.Start(ctx, "OrderService.CreateOrder",
		trace.WithAttributes(attribute.Int64("user.id", int64(userId))))
	defer span.End()

	var orderResponse dto.OrderResponse

	err := o.db.Transaction(func(tx *gorm.DB) error {
		// Bind transactional scope repositories safely
		cartRepository := o.cartRepository.WithTx(tx)
		cartItemRepository := o.cartItemRepository.WithTx(tx)
		productRepository := o.productRepository.WithTx(tx)
		orderRepository := o.orderRepository.WithTx(tx)

		// Fetches cart while preloading active product discounts matching the UTC timeline
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
			product := &item.Product

			// Verify inventory availability
			if product.Stock < item.Quantity {
				return fmt.Errorf("not enough stock for product: %s", product.Name)
			}

			// Decrement stock allocation and save changes inside transaction block
			product.Stock -= item.Quantity
			if err := productRepository.UpdateProduct(ctx, product); err != nil {
				return err
			}

			// Determine the actual sale unit price for this product
			actualUnitPrice := product.Price
			if len(product.Discounts) > 0 {
				actualUnitPrice = CalculateDiscountPrice(&product.Discounts[0], product.Price)
			}

			// Compute total amount using the accurate price
			totalAmount += float64(item.Quantity) * actualUnitPrice

			orderItems = append(orderItems, domain.OrderItem{
				ProductId: product.Id,
				Quantity:  item.Quantity,
				Price:     actualUnitPrice, // Persist the actual price paid into the history ledger
			})
		}

		// Construct the pure Order model without tracking coupon strings
		order := &domain.Order{
			UserId:      userId,
			Status:      domain.OrderStatusPending,
			TotalAmount: totalAmount,
			OrderItems:  orderItems,
		}

		if err := orderRepository.CreateOrder(ctx, order); err != nil {
			return fmt.Errorf("orderRepository.CreateOrder: %w", err)
		}

		span.SetAttributes(
			attribute.Int64("order.id", int64(order.Id)),
			attribute.Float64("order.total", totalAmount),
		)

		// Clear items out of user's basket
		if err := cartItemRepository.ClearCartItems(ctx, cart.Id); err != nil {
			return fmt.Errorf("cartItemRepository.ClearCartItems: %w", err)
		}

		// Reload order graph completely (pulling database relationships) to map DTO fields safely
		createdOrder, err := orderRepository.GetOrderById(ctx, order.Id)
		if err != nil {
			return fmt.Errorf("orderRepository.GetOrderById: %w", err)
		}

		resp := o.toOrderResp(createdOrder)
		orderResponse = *resp
		return nil
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create order")
		return nil, err
	}

	return &orderResponse, nil
}

func (o *orderService) GetOrder(ctx context.Context, userId, orderId uint) (*dto.OrderResponse, error) {
	ctx, span := o.tracer.Start(ctx, "OrderService.GetOrder",
		trace.WithAttributes(
			attribute.Int64("user.id", int64(userId)),
			attribute.Int64("order.id", int64(orderId)),
		))
	defer span.End()

	order, err := o.orderRepository.GetOrderByUserId(ctx, userId, orderId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "order not found")
		return nil, fmt.Errorf("o.orderRepository.GetOrderByUserId: %w", err)
	}

	return o.toOrderResp(order), nil
}

func (o *orderService) GetOrders(ctx context.Context, userId uint, page, limit int) ([]*dto.OrderResponse, *helper.PaginatedMeta, error) {
	ctx, span := o.tracer.Start(ctx, "OrderService.GetOrders",
		trace.WithAttributes(
			attribute.Int64("user.id", int64(userId)),
			attribute.Int("page", page),
			attribute.Int("limit", limit),
		))
	defer span.End()

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
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to count orders")
		return nil, nil, fmt.Errorf("o.orderRepository.CountUserOrders: %w", err)
	}

	orders, err := o.orderRepository.GetUserOrders(ctx, userId, offset, limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get orders")
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
		product := &item.Product

		isOnSale := false
		discountedPrice := product.Price

		if len(product.Discounts) > 0 {
			calculatedPrice := CalculateDiscountPrice(&product.Discounts[0], product.Price)
			if calculatedPrice < product.Price {
				discountedPrice = calculatedPrice
				isOnSale = true
			}
		}

		orderItems[i] = dto.OrderItemResponse{
			Id: item.Id,
			Product: dto.ProductResponse{
				Id:              product.Id,
				CategoryId:      product.CategoryId,
				Name:            product.Name,
				Description:     product.Description,
				Price:           product.Price,
				IsOnSale:        isOnSale,
				DiscountedPrice: discountedPrice,
				Stock:           product.Stock,
				SKU:             product.SKU,
				IsActive:        product.IsActive,
				Category: dto.CategoryResponse{
					Id:          product.Category.Id,
					Name:        product.Category.Name,
					Description: product.Category.Description,
					IsActive:    product.Category.IsActive,
					CreatedAt:   product.Category.CreatedAt, // FIX: Map category created_at timestamp
					UpdatedAt:   product.Category.UpdatedAt, // FIX: Map category updated_at timestamp
				},
				Images:    make([]dto.ProductImageResponse, 0),
				CreatedAt: product.CreatedAt, // FIX: Map product created_at timestamp
				UpdatedAt: product.UpdatedAt, // FIX: Map product updated_at timestamp
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

func NewOrderService(orderRepository repository.OrderRepository, cartRepository repository.CartRepository, cartItemRepository repository.CartItemRepository, productRepository repository.ProductRepository, db *gorm.DB, tracer trace.Tracer) OrderService {
	return &orderService{
		orderRepository:    orderRepository,
		cartRepository:     cartRepository,
		cartItemRepository: cartItemRepository,
		productRepository:  productRepository,
		db:                 db,
		tracer:             tracer,
	}
}
