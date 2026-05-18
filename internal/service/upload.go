package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/saleh-ghazimoradi/GopherMarket/pkg/uploadStrategy"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"io"
	"path/filepath"
	"strings"
)

type UploadService interface {
	UploadProductImage(ctx context.Context, productId uint, file io.Reader, filename string) (string, error)
}

type uploadService struct {
	uploadStrategy uploadStrategy.UploadStrategy
	tracer         trace.Tracer
}

func (u *uploadService) UploadProductImage(ctx context.Context, productId uint, file io.Reader, filename string) (string, error) {
	ctx, span := u.tracer.Start(ctx, "UploadService.UploadProductImage",
		trace.WithAttributes(
			attribute.Int64("product.id", int64(productId)),
			attribute.String("filename", filename),
		))
	defer span.End()

	ext := strings.ToLower(filepath.Ext(filename))
	if !isValidImageExt(ext) {
		span.SetStatus(codes.Error, "invalid file type")
		return "", fmt.Errorf("invalid file type: %s", ext)
	}

	newFileName := uuid.New().String()
	path := fmt.Sprintf("products/%d/%s%s", productId, newFileName, ext)

	url, err := u.uploadStrategy.UploadFile(file, filename, path)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "upload failed")
		return "", err
	}

	return url, nil
}

func isValidImageExt(ext string) bool {
	validExtS := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	for _, validExt := range validExtS {
		if ext == validExt {
			return true
		}
	}
	return false
}

func NewUploadService(uploadStrategy uploadStrategy.UploadStrategy, tracer trace.Tracer) UploadService {
	return &uploadService{
		uploadStrategy: uploadStrategy,
		tracer:         tracer,
	}
}
