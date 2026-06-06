package service

import (
	"context"
	"errors"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type CartService interface {
	GetCart(ctx context.Context, userId uint) (*dto.CartResponse, error)
	AddToCart(ctx context.Context, userId uint, req *dto.AddToCartRequest) (*dto.CartResponse, error)
	UpdateCartItem(ctx context.Context, userId, itemId uint, req *dto.UpdateCartItemRequest) (*dto.CartResponse, error)
	RemoveFromCart(ctx context.Context, userId, itemId uint) error
}

type cartService struct {
	cartRepository     repository.CartRepository
	cartItemRepository repository.CartItemRepository
	productRepository  repository.ProductRepository
	tracer             trace.Tracer
}

func (c *cartService) GetCart(ctx context.Context, userId uint) (*dto.CartResponse, error) {
	ctx, span := c.tracer.Start(ctx, "CartService.GetCart",
		trace.WithAttributes(attribute.Int64("user.id", int64(userId))))
	defer span.End()

	cart, err := c.cartRepository.GetCartByUserId(ctx, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get cart")
		return nil, err
	}

	return c.convertToCartResponse(cart), nil
}

func (c *cartService) AddToCart(ctx context.Context, userId uint, req *dto.AddToCartRequest) (*dto.CartResponse, error) {
	ctx, span := c.tracer.Start(ctx, "CartService.AddToCart",
		trace.WithAttributes(
			attribute.Int64("user.id", int64(userId)),
			attribute.Int64("product.id", int64(req.ProductId)),
			attribute.Int("product.quantity", req.Quantity),
		))
	defer span.End()

	product, err := c.productRepository.GetProductById(ctx, req.ProductId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "product not found")
		return nil, errors.New("product not found")
	}

	if product.Stock < req.Quantity {
		span.SetStatus(codes.Error, "not enough stock")
		return nil, errors.New("not enough stock")
	}

	cart, err := c.cartRepository.GetOrCreateCart(ctx, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get or create cart")
		return nil, err
	}

	cartItem, err := c.cartItemRepository.GetCartItem(ctx, cart.Id, req.ProductId)
	if err != nil {
		item := &domain.CartItem{
			CartId:    cart.Id,
			ProductId: product.Id,
			Quantity:  req.Quantity,
		}
		_ = c.cartItemRepository.CreateCartItem(ctx, item)
	} else {
		cartItem.Quantity += req.Quantity
		if cartItem.Quantity > product.Stock {
			span.SetStatus(codes.Error, "not enough stock")
			return nil, errors.New("not enough stock")
		}
		_ = c.cartItemRepository.UpdateCartItem(ctx, cartItem)
	}

	return c.GetCart(ctx, userId)
}

func (c *cartService) UpdateCartItem(ctx context.Context, userId, itemId uint, req *dto.UpdateCartItemRequest) (*dto.CartResponse, error) {
	ctx, span := c.tracer.Start(ctx, "CartService.UpdateCartItem",
		trace.WithAttributes(
			attribute.Int64("user.id", int64(userId)),
			attribute.Int64("item.id", int64(itemId)),
		))
	defer span.End()

	cartItem, err := c.cartItemRepository.GetCartItemWithUser(ctx, userId, itemId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cart item not found")
		return nil, err
	}

	product, err := c.productRepository.GetProductById(ctx, cartItem.ProductId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "product not found")
		return nil, err
	}

	if product.Stock < req.Quantity {
		span.SetStatus(codes.Error, "not enough stock")
		return nil, errors.New("not enough stock")
	}

	cartItem.Quantity = req.Quantity
	if err := c.cartItemRepository.UpdateCartItem(ctx, cartItem); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update cart item")
		return nil, err
	}

	return c.GetCart(ctx, userId)
}

func (c *cartService) RemoveFromCart(ctx context.Context, userId, itemId uint) error {
	ctx, span := c.tracer.Start(ctx, "CartService.RemoveFromCart",
		trace.WithAttributes(
			attribute.Int64("user.id", int64(userId)),
			attribute.Int64("item.id", int64(itemId)),
		))
	defer span.End()

	if err := c.cartItemRepository.DeleteCartItem(ctx, userId, itemId); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to remove item")
		return err
	}
	return nil
}

func (c *cartService) convertToCartResponse(cart *domain.Cart) *dto.CartResponse {
	cartItems := make([]dto.CartItemResponse, len(cart.CartItems))
	var total float64

	for i := range cart.CartItems {
		item := &cart.CartItems[i]
		product := &item.Product

		sellingPrice := product.Price
		isOnSale := false

		if len(product.Discounts) > 0 {
			calculatedPrice := CalculateDiscountPrice(&product.Discounts[0], product.Price)
			if calculatedPrice < product.Price {
				sellingPrice = calculatedPrice
				isOnSale = true
			}
		}

		subtotal := float64(item.Quantity) * sellingPrice
		total += subtotal

		images := make([]dto.ProductImageResponse, len(product.Images))
		for j := range product.Images {
			images[j] = dto.ProductImageResponse{
				Id:        product.Images[j].Id,
				URL:       product.Images[j].URL,
				AltText:   product.Images[j].AltText,
				IsPrimary: product.Images[j].IsPrimary,
				CreatedAt: product.Images[j].CreatedAt,
			}
		}

		cartItems[i] = dto.CartItemResponse{
			Id: item.Id,
			Product: dto.ProductResponse{
				Id:              product.Id,
				CategoryId:      product.CategoryId,
				Name:            product.Name,
				Description:     product.Description,
				Price:           product.Price,
				IsOnSale:        isOnSale,
				DiscountedPrice: sellingPrice,
				Stock:           product.Stock,
				SKU:             product.SKU,
				IsActive:        product.IsActive,
				Category: dto.CategoryResponse{
					Id:          product.Category.Id,
					Name:        product.Category.Name,
					Description: product.Category.Description,
					IsActive:    product.Category.IsActive,
					CreatedAt:   product.Category.CreatedAt,
					UpdatedAt:   product.Category.UpdatedAt,
				},
				Images:    images,
				CreatedAt: product.CreatedAt,
				UpdatedAt: product.UpdatedAt,
			},
			Quantity:  item.Quantity,
			Subtotal:  subtotal,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
	}

	return &dto.CartResponse{
		Id:        cart.Id,
		UserId:    cart.UserId,
		CartItems: cartItems,
		Total:     total,
		CreatedAt: cart.CreatedAt,
		UpdatedAt: cart.UpdatedAt,
	}
}

func NewCartService(cartRepository repository.CartRepository, cartItemRepository repository.CartItemRepository, productRepository repository.ProductRepository, tracer trace.Tracer) CartService {
	return &cartService{
		cartRepository:     cartRepository,
		cartItemRepository: cartItemRepository,
		productRepository:  productRepository,
		tracer:             tracer,
	}
}
