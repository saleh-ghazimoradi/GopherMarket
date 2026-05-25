package middleware

import (
	"context"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/saleh-ghazimoradi/GopherMarket/utils"
	"net/http"
	"strings"
)

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization == "" {
			helper.UnauthorizedResponse(w, "Authorization header missing")
			return
		}

		tokenParts := strings.Split(authorization, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			helper.UnauthorizedResponse(w, "Invalid authorization header format")
			return
		}

		claims, err := utils.ValidateToken(tokenParts[1], m.cfg.JWT.Secret)
		if err != nil {
			helper.UnauthorizedResponse(w, "Unauthorized")
			return
		}

		ctx := r.Context()
		ctx = utils.WithUserId(ctx, claims.UserId)
		ctx = utils.WithEmailKey(ctx, claims.Email)
		ctx = utils.WithRoleKey(ctx, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Admin ensures the authenticated user has the admin role.
func (m *Middleware) Admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, exists := utils.RoleFromContext(r.Context())
		if !exists || role != string(domain.Admin) {
			helper.ForbiddenResponse(w, "Access Denied")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GraphQLAuth injects user context into GraphQL requests, but does NOT block unauthenticated ones.
func (m *Middleware) GraphQLAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), HTTPResponseKey, w)
		ctx = context.WithValue(ctx, HTTPRequestKey, r)
		r = r.WithContext(ctx)

		authorization := r.Header.Get("Authorization")
		if authorization != "" {
			tokenParts := strings.Split(authorization, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				if claims, err := utils.ValidateToken(tokenParts[1], m.cfg.JWT.Secret); err == nil {
					ctx = r.Context()
					ctx = utils.WithUserId(ctx, claims.UserId)
					ctx = utils.WithEmailKey(ctx, claims.Email)
					ctx = utils.WithRoleKey(ctx, claims.Role)
					r = r.WithContext(ctx)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// WrapAuth is a convenience method that wraps a single http.HandlerFunc with Authenticate.
func (m *Middleware) WrapAuth(handler http.HandlerFunc) http.Handler {
	return m.Authenticate(http.HandlerFunc(handler))
}

// WrapAdmin wraps a handler with both Authenticate and Admin.
func (m *Middleware) WrapAdmin(handler http.HandlerFunc) http.Handler {
	return m.Authenticate(m.Admin(http.HandlerFunc(handler)))
}
