package middleware

import (
	"net/http"
	"strings"
)

type HPPOptions struct {
	CheckQuery                  bool
	CheckBody                   bool
	CheckBodyOnlyForContentType string
	Whitelist                   []string
}

func (m *Middleware) HPP(options HPPOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if options.CheckBody && r.Method == http.MethodPost && m.isCorrectContentType(r, options.CheckBodyOnlyForContentType) {
				m.filterBodyParams(r, options.Whitelist)
			}
			if options.CheckQuery && r.URL.Query() != nil {
				m.filterQueryParams(r, options.Whitelist)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) isCorrectContentType(r *http.Request, contentType string) bool {
	return strings.Contains(r.Header.Get("Content-Type"), contentType)
}

func (m *Middleware) filterBodyParams(r *http.Request, whitelist []string) {
	if err := r.ParseForm(); err != nil {
		m.logger.Error("filterBodyParams", "err", err)
		return
	}
	for k, v := range r.Form {
		if len(v) > 1 {
			r.Form.Set(k, v[0])
		}
		if !m.isWhiteListed(k, whitelist) {
			delete(r.Form, k)
		}
	}
}

func (m *Middleware) filterQueryParams(r *http.Request, whitelist []string) {
	query := r.URL.Query()
	for k, v := range query {
		if len(v) > 1 {
			query.Set(k, v[0])
		}
		if !m.isWhiteListed(k, whitelist) {
			query.Del(k)
		}
	}
	r.URL.RawQuery = query.Encode()
}

func (m *Middleware) isWhiteListed(param string, whitelist []string) bool {
	for _, v := range whitelist {
		if v == param {
			return true
		}
	}
	return false
}
