package uploadStrategy

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
)

type LocalStrategy struct {
	basePath string
}

func (l *LocalStrategy) UploadFile(file *multipart.FileHeader, path string) (string, error) {
	fullPath := filepath.Join(l.basePath, path)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", fmt.Errorf("mkdir err: %v", err)
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open err: %v", err)
	}

	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create err: %v", err)
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(src); err != nil {
		return "", fmt.Errorf("read err: %v", err)
	}

	return fmt.Sprintf("/uploads/%s", path), nil
}

func (l *LocalStrategy) DeleteFile(path string) error {
	fullPath := filepath.Join(l.basePath, path)
	return os.Remove(fullPath)
}

func NewLocalStrategy(basePath string) *LocalStrategy {
	return &LocalStrategy{
		basePath: basePath,
	}
}
