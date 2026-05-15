package middleware

import (
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/saleh-ghazimoradi/GopherMarket/utils"
	"log/slog"
	"net/http"
	"strings"
)

type Middleware struct {
	logger *slog.Logger
	cfg    *config.Config
}

func (m *Middleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logger.Info("Incoming request: ", "method", r.Method, "path", r.URL.Path, "protocol", r.Proto, "remote_addr", r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				helper.InternalServerError(w, "panic recovery hit", fmt.Errorf("%v", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

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
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)

	})
}

func (m *Middleware) Admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, exists := utils.RoleFromContext(r.Context())
		if !exists {
			helper.ForbiddenResponse(w, "Access Denied")
			return
		}

		if role != string(domain.Admin) {
			helper.ForbiddenResponse(w, "Access Denied")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) GraphQLAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization != "" {
			tokenParts := strings.Split(authorization, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				if claims, err := utils.ValidateToken(tokenParts[1], m.cfg.JWT.Secret); err == nil {
					ctx := r.Context()
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

func (m *Middleware) WrapAuth(handlerFunc http.HandlerFunc) http.Handler {
	return m.Authenticate(handlerFunc)
}

func (m *Middleware) WrapAdmin(handlerFunc http.HandlerFunc) http.Handler {
	return m.Authenticate(m.Admin(handlerFunc))
}

func NewMiddleware(logger *slog.Logger, cfg *config.Config) *Middleware {
	return &Middleware{
		logger: logger,
		cfg:    cfg,
	}
}
