package service

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/saleh-ghazimoradi/GopherMarket/pkg/uploadStrategy"
	"mime/multipart"
	"path/filepath"
	"strings"
)

type UploadService interface {
	UploadProductImage(productId uint, file *multipart.FileHeader) (string, error)
}

type uploadService struct {
	uploadStrategy uploadStrategy.UploadStrategy
}

func (u *uploadService) UploadProductImage(productId uint, file *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))

	if !isValidImageExt(ext) {
		return "", fmt.Errorf("invalid file type: %s", ext)
	}

	newFileName := uuid.New().String()

	path := fmt.Sprintf("products/%d/%s%s", productId, newFileName, ext)

	return u.uploadStrategy.UploadFile(file, path)
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

func NewUploadService(uploadStrategy uploadStrategy.UploadStrategy) UploadService {
	return &uploadService{
		uploadStrategy: uploadStrategy,
	}
}
