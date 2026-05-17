package uploadStrategy

import "io"

type UploadStrategy interface {
	UploadFile(file io.Reader, filename, path string) (string, error)
	DeleteFile(path string) error
}
