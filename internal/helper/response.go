package helper

import "net/http"

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Error   string `json:"error"`
}

type PaginatedResponse struct {
	Response Response
	Meta     PaginatedMeta `json:"meta"`
}

type PaginatedMeta struct {
	Page      int   `json:"page"`
	Limit     int   `json:"limit"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"total_page"`
}

func SuccessResponse(w http.ResponseWriter, message string, data any) {
	response := Response{
		Success: true,
		Message: message,
		Data:    data,
	}
	writeJSON(w, http.StatusOK, response)
}

func CreatedResponse(w http.ResponseWriter, message string, data any) {
	response := Response{
		Success: true,
		Message: message,
		Data:    data,
	}
	writeJSON(w, http.StatusCreated, response)
}

func ErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	response := Response{
		Success: false,
		Message: message,
	}
	if err != nil {
		response.Error = err.Error()
	}
	writeJSON(w, statusCode, response)
}

func BadRequestResponse(w http.ResponseWriter, message string, err error) {
	ErrorResponse(w, http.StatusBadRequest, message, err)
}

func UnauthorizedResponse(w http.ResponseWriter, message string) {
	ErrorResponse(w, http.StatusUnauthorized, message, nil)
}

func ForbiddenResponse(w http.ResponseWriter, message string) {
	ErrorResponse(w, http.StatusForbidden, message, nil)
}

func NotFoundResponse(w http.ResponseWriter, message string) {
	ErrorResponse(w, http.StatusNotFound, message, nil)
}

func InternalServerError(w http.ResponseWriter, message string, err error) {
	ErrorResponse(w, http.StatusInternalServerError, message, err)
}

func FailedValidationResponse(w http.ResponseWriter, message string) {
	ErrorResponse(w, http.StatusUnprocessableEntity, message, nil)
}

func EditConflictResponse(w http.ResponseWriter, message string, err error) {
	ErrorResponse(w, http.StatusConflict, message, err)
}

func RateLimitExceededResponse(w http.ResponseWriter, message string) {
	ErrorResponse(w, http.StatusTooManyRequests, message, nil)
}

func PaginatedSuccessResponse(w http.ResponseWriter, message string, data any, meta PaginatedMeta) {
	paginatedResponse := PaginatedResponse{
		Response: Response{
			Success: true,
			Message: message,
			Data:    data,
		},
		Meta: meta,
	}
	writeJSON(w, http.StatusOK, paginatedResponse)
}
