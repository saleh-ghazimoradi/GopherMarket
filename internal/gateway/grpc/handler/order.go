package handler

import (
	"context"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/grpc/protos"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderGrpcHandler struct {
	protos.UnimplementedOrderServiceServer
	orderService service.OrderService
}

func (h *OrderGrpcHandler) CreateOrder(ctx context.Context, req *protos.CreateOrderRequest) (*protos.OrderResponse, error) {
	resp, err := h.orderService.CreateOrder(ctx, uint(req.UserId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create order transaction: %v", err)
	}

	return h.mapToOrderResponse(resp), nil
}

func (h *OrderGrpcHandler) GetOrder(ctx context.Context, req *protos.GetOrderRequest) (*protos.OrderResponse, error) {
	resp, err := h.orderService.GetOrder(ctx, uint(req.UserId), uint(req.OrderId))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to retrieve order records: %v", err)
	}

	return h.mapToOrderResponse(resp), nil
}

func (h *OrderGrpcHandler) GetOrders(ctx context.Context, req *protos.GetOrdersRequest) (*protos.GetOrdersResponse, error) {
	orders, meta, err := h.orderService.GetOrders(ctx, uint(req.UserId), int(req.Page), int(req.Limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve paginated orders matching criteria: %v", err)
	}

	protoOrders := make([]*protos.OrderResponse, len(orders))
	for i, ord := range orders {
		protoOrders[i] = h.mapToOrderResponse(ord)
	}

	return &protos.GetOrdersResponse{
		Orders: protoOrders,
		Meta:   h.mapToPaginatedMeta(meta),
	}, nil
}

func (h *OrderGrpcHandler) mapToOrderResponse(ord *dto.OrderResponse) *protos.OrderResponse {
	if ord == nil {
		return nil
	}

	protoItems := make([]*protos.OrderItemResponse, len(ord.OrderItems))
	for i, item := range ord.OrderItems {
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

		protoItems[i] = &protos.OrderItemResponse{
			Id:       uint64(item.Id),
			Quantity: int32(item.Quantity),
			Price:    item.Price,
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
		}
	}

	return &protos.OrderResponse{
		Id:          uint64(ord.Id),
		UserId:      uint64(ord.UserId),
		Status:      ord.Status,
		TotalAmount: ord.TotalAmount,
		OrderItems:  protoItems,
		CreatedAt:   timestamppb.New(ord.CreatedAt),
		UpdatedAt:   timestamppb.New(ord.UpdatedAt),
	}
}

func (h *OrderGrpcHandler) mapToPaginatedMeta(m *helper.PaginatedMeta) *protos.PaginatedMeta {
	if m == nil {
		return nil
	}
	return &protos.PaginatedMeta{
		Page:       int32(m.Page),
		Limit:      int32(m.Limit),
		Total:      m.Total,
		TotalPages: int32(m.TotalPage),
	}
}

func NewOrderGrpcHandler(orderService service.OrderService) *OrderGrpcHandler {
	return &OrderGrpcHandler{
		orderService: orderService,
	}
}
