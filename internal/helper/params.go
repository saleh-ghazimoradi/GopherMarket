package helper

import (
	"net/http"
	"strconv"
)

func ReadParams(r *http.Request, name string) (uint, error) {
	id := r.PathValue(name)
	uintId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(uintId), nil
}

func ReadQueryParam(r *http.Request, key string) (int, error) {
	query, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil {
		return 0, err
	}
	return query, nil
}
