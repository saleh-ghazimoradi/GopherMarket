package middleware

type contextKey string

const (
	HTTPRequestKey  contextKey = "http_request"
	HTTPResponseKey contextKey = "http_response"
)
