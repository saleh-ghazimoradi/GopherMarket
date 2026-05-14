package resolver

import (
	"context"
	"errors"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/utils"
)

func getPagingNumbers(page, limit *int) (int, int) {
	p, l := 0, 0

	if page != nil {
		p = *page
	}

	if limit != nil {
		l = *limit
	}

	if p <= 0 {
		p = 1
	}

	if l <= 0 {
		l = 10
	}

	return p, l
}

func isAdmin(ctx context.Context) (bool, error) {
	role, exists := utils.RoleFromContext(ctx)
	if !exists {
		return false, errors.New("unauthorized")
	}

	return role == string(domain.Admin), nil
}
