package middleware

//import (
//	"bytes"
//	"fmt"
//	"github.com/microcosm-cc/bluemonday"
//	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
//	"io"
//	"net/http"
//	"strings"
//)
//
//func (m *Middleware) XSS(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		cleanedPath := m.sanitizeString(r.URL.Path)
//		r.URL.Path = cleanedPath
//
//		query := r.URL.Query()
//		for key, values := range query {
//			query.Del(key)
//			for _, val := range values {
//				query.Add(m.sanitizeString(key), m.sanitizeString(val))
//			}
//		}
//		r.URL.RawQuery = query.Encode()
//
//		contentType := r.Header.Get("Content-Type")
//		if r.Body != nil && strings.HasPrefix(contentType, "application/json") {
//			bodyBytes, err := io.ReadAll(r.Body)
//			if err != nil {
//				helper.BadRequestResponse(w, "Failed to read request body", err)
//				return
//			}
//			r.Body.Close()
//
//			if len(bodyBytes) > 0 {
//				sanitizedBody := m.sanitizeString(string(bodyBytes))
//				r.Body = io.NopCloser(bytes.NewReader([]byte(sanitizedBody)))
//			} else {
//				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
//			}
//		}
//
//		next.ServeHTTP(w, r)
//	})
//}
//
//// Clean sanitizes input data to prevent XSS attacks
//func (m *Middleware) clean(data interface{}) (interface{}, error) {
//	switch v := data.(type) {
//	case map[string]interface{}:
//		for key, value := range v {
//			v[key] = m.sanitizeValue(value)
//		}
//		return v, nil
//	case []interface{}:
//		for i, value := range v {
//			v[i] = m.sanitizeValue(value)
//		}
//		return v, nil
//	case string:
//		return m.sanitizeString(v), nil
//	default:
//		return nil, fmt.Errorf("invalid type %T", v)
//	}
//}
//
//func (m *Middleware) sanitizeValue(data interface{}) interface{} {
//	switch v := data.(type) {
//	case string:
//		return m.sanitizeString(v)
//	case map[string]interface{}:
//		for k, value := range v {
//			v[k] = m.sanitizeValue(value)
//		}
//		return v
//	case []interface{}:
//		for i, value := range v {
//			v[i] = m.sanitizeValue(value)
//		}
//		return v
//	default:
//		return v // Return v as it is unsupported
//	}
//}
//
//func (m *Middleware) sanitizeString(value string) string {
//	return bluemonday.UGCPolicy().Sanitize(value)
//}
