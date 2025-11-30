package util

import (
	"crypto/rand"
	"math/big"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const idLength = 8

// GenerateID creates a random 8-character alphanumeric ID
func GenerateID() (string, error) {
	id := make([]byte, idLength)
	for i := range id {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		id[i] = charset[n.Int64()]
	}
	return string(id), nil
}
