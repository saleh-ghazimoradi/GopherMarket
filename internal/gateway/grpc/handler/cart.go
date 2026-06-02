package handler

import (
	"context"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/grpc/protos"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CartGrpcHandler struct {
	protos.UnimplementedCartServiceServer
	cartService service.CartService
}

func (c *CartGrpcHandler) GetCart(ctx context.Context, req *protos.GetCartRequest) (*protos.CartResponse, error) {
	resp, err := c.cartService.GetCart(ctx, uint(req.UserId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get cart: %v", err)
	}

	return c.mapToCartResponse(resp), nil
}

func (c *CartGrpcHandler) AddToCart(ctx context.Context, req *protos.AddToCartGrpcRequest) (*protos.CartResponse, error) {
	if req.Request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "missing item payload specifications")
	}

	actionDTO := &dto.AddToCartRequest{
		ProductId: uint(req.Request.ProductId),
		Quantity:  int(req.Request.Quantity),
	}

	resp, err := c.cartService.AddToCart(ctx, uint(req.UserId), actionDTO)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "add operation aborted: %v", err)
	}

	return c.mapToCartResponse(resp), nil
}

func (c *CartGrpcHandler) UpdateCartItem(ctx context.Context, req *protos.UpdateCartItemGrpcRequest) (*protos.CartResponse, error) {
	if req.Request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "missing modification updates payload specifications")
	}

	updateDTO := &dto.UpdateCartItemRequest{
		Quantity: int(req.Request.Quantity),
	}

	resp, err := c.cartService.UpdateCartItem(ctx, uint(req.UserId), uint(req.ItemId), updateDTO)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update item quantity: %v", err)
	}

	return c.mapToCartResponse(resp), nil
}

func (c *CartGrpcHandler) RemoveFromCart(ctx context.Context, req *protos.RemoveFromCartRequest) (*emptypb.Empty, error) {
	err := c.cartService.RemoveFromCart(ctx, uint(req.UserId), uint(req.ItemId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to drop item from cart tracking: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (c *CartGrpcHandler) mapToCartResponse(cart *dto.CartResponse) *protos.CartResponse {
	if cart == nil {
		return nil
	}

	protoItems := make([]*protos.CartItemResponse, len(cart.CartItems))
	for i, item := range cart.CartItems {
		protoImages := make([]*protos.ProductImageResponse, len(item.Product.Images))
		for j, img := range item.Product.Images {
			protoImages[j] = &protos.ProductImageResponse{
				Id:        uint64(img.Id),
				Url:       img.URL,
				AltText:   img.AltText,
				IsPrimary: img.IsPrimary,
				CreatedAt: timestamppb.New(img.CreatedAt),
			}
		}

		protoItems[i] = &protos.CartItemResponse{
			Id:       uint64(item.Id),
			Quantity: int32(item.Quantity),
			Subtotal: item.Subtotal,
			Product: &protos.ProductResponse{
				Id:          uint64(item.Product.Id),
				CategoryId:  uint64(item.Product.CategoryId),
				Name:        item.Product.Name,
				Description: item.Product.Description,
				Price:       item.Product.Price,
				Stock:       int32(item.Product.Stock),
				Sku:         item.Product.SKU,
				IsActive:    item.Product.IsActive,
				Category: &protos.CategoryResponse{
					Id:          uint64(item.Product.Category.Id),
					Name:        item.Product.Category.Name,
					Description: item.Product.Category.Description,
					IsActive:    item.Product.Category.IsActive,
					CreatedAt:   timestamppb.New(item.Product.Category.CreatedAt),
					UpdatedAt:   timestamppb.New(item.Product.Category.UpdatedAt),
				},
				Images:    protoImages,
				CreatedAt: timestamppb.New(item.Product.CreatedAt),
				UpdatedAt: timestamppb.New(item.Product.UpdatedAt),
			},
			CreatedAt: timestamppb.New(item.CreatedAt),
			UpdatedAt: timestamppb.New(item.UpdatedAt),
		}
	}

	return &protos.CartResponse{
		Id:        uint64(cart.Id),
		UserId:    uint64(cart.UserId),
		Total:     cart.Total,
		CartItems: protoItems,
		CreatedAt: timestamppb.New(cart.CreatedAt),
		UpdatedAt: timestamppb.New(cart.UpdatedAt),
	}
}

func NewCartGrpcHandler(cartService service.CartService) *CartGrpcHandler {
	return &CartGrpcHandler{
		cartService: cartService,
	}
}
