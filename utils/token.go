package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateHexToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	return token, nil
}
