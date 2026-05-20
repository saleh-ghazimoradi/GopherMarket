package utils

import (
	"crypto/rand"
	"math/big"
)

const codeCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func GenerateSecureCode(length int) (string, error) {
	if length <= 0 {
		length = 8
	}

	result := make([]byte, length)
	for i := range result {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(codeCharset))))
		if err != nil {
			return "", err
		}
		result[i] = codeCharset[idx.Int64()]
	}
	return string(result), nil
}
