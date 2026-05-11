package uploadStrategy

import "mime/multipart"

type UploadStrategy interface {
	UploadFile(file *multipart.FileHeader, path string) (string, error)
	DeleteFile(path string) error
}
