package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// GenerateRandomToken returns a cryptographically secure random token with size bytes.
func GenerateRandomToken(size int) (string, error) {
	if size <= 0 {
		return "", fmt.Errorf("token size must be positive")
	}

	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return hex.EncodeToString(buffer), nil
}

// GenerateNumericCode returns a string of numeric digits with the requested length.
func GenerateNumericCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("code length must be positive")
	}

	var code strings.Builder
	max := big.NewInt(10)

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		code.WriteByte('0' + byte(n.Int64()))
	}

	return code.String(), nil
}
