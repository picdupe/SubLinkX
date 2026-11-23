package utils

import (
	"crypto/rand"
	"errors"
	"math/big"
)

// Generate a fixed-length string using cryptographically secure randomness.
func RandString(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("random string length must be positive")
	}
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	limit := big.NewInt(int64(len(alphabet)))
	result := make([]byte, length)
	for i := range result {
		index, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", err
		}
		result[i] = alphabet[index.Int64()]
	}
	return string(result), nil
}
