package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/repository"
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
}

func (c *cartService) GetCart(ctx context.Context, userId uint) (*dto.CartResponse, error) {
	cart, err := c.cartRepository.GetCartByUserId(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("get cart by user id %w", err)
	}

	return c.toCartReps(cart), nil
}

func (c *cartService) AddToCart(ctx context.Context, userId uint, req *dto.AddToCartRequest) (*dto.CartResponse, error) {
	product, err := c.productRepository.GetProductById(ctx, req.ProductId)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if product.Stock < req.Quantity {
		return nil, errors.New("not enough stock")
	}

	cart, err := c.cartRepository.GetOrCreateCart(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("get or create cart %w", err)
	}

	cartItem, err := c.cartItemRepository.GetCartItem(ctx, cart.Id, product.Id)
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
			return nil, errors.New("not enough stock")
		}

		_ = c.cartItemRepository.UpdateCartItem(ctx, cartItem)
	}

	return c.GetCart(ctx, userId)
}

func (c *cartService) UpdateCartItem(ctx context.Context, userId, itemId uint, req *dto.UpdateCartItemRequest) (*dto.CartResponse, error) {
	cartItem, err := c.cartItemRepository.GetCartItemWithUser(ctx, userId, itemId)
	if err != nil {
		return nil, fmt.Errorf("get cart item with user id %w", err)
	}

	product, err := c.productRepository.GetProductById(ctx, cartItem.ProductId)
	if err != nil {
		return nil, fmt.Errorf("get product id %w", err)
	}

	if product.Stock < req.Quantity {
		return nil, errors.New("not enough stock")
	}

	cartItem.Quantity = req.Quantity
	if err := c.cartItemRepository.UpdateCartItem(ctx, cartItem); err != nil {
		return nil, fmt.Errorf("update cart item %w", err)
	}

	return c.GetCart(ctx, userId)
}

func (c *cartService) RemoveFromCart(ctx context.Context, userId, itemId uint) error {
	return c.cartItemRepository.DeleteCartItem(ctx, userId, itemId)
}

func (c *cartService) toCartReps(cart *domain.Cart) *dto.CartResponse {
	cartItems := make([]dto.CartItemResponse, len(cart.CartItems))
	var total float64

	for i := range cart.CartItems {
		subtotal := float64(cart.CartItems[i].Quantity) * cart.CartItems[i].Product.Price
		total += subtotal

		cartItems[i] = dto.CartItemResponse{
			Id: cart.CartItems[i].Id,
			Product: dto.ProductResponse{
				Id:          cart.CartItems[i].Product.Id,
				CategoryId:  cart.CartItems[i].Product.CategoryId,
				Name:        cart.CartItems[i].Product.Name,
				Description: cart.CartItems[i].Product.Description,
				Price:       cart.CartItems[i].Product.Price,
				Stock:       cart.CartItems[i].Product.Stock,
				SKU:         cart.CartItems[i].Product.SKU,
				IsActive:    cart.CartItems[i].Product.IsActive,
				Category: dto.CategoryResponse{
					Id:          cart.CartItems[i].Product.Category.Id,
					Name:        cart.CartItems[i].Product.Name,
					Description: cart.CartItems[i].Product.Description,
					IsActive:    cart.CartItems[i].Product.IsActive,
				},
			},
			Quantity:  cart.CartItems[i].Quantity,
			Subtotal:  subtotal,
			CreatedAt: cart.CartItems[i].CreatedAt,
			UpdatedAt: cart.CartItems[i].UpdatedAt,
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

func NewCartService(cartRepository repository.CartRepository, cartItemRepository repository.CartItemRepository, productRepository repository.ProductRepository) CartService {
	return &cartService{
		cartRepository:     cartRepository,
		cartItemRepository: cartItemRepository,
		productRepository:  productRepository,
	}
}
