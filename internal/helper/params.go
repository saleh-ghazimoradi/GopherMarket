package helper

import (
	"net/http"
	"strconv"
)

func ReadParams(r *http.Request) (uint, error) {
	id := r.PathValue("id")
	uintId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(uintId), nil
}
