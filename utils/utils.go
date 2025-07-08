package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateID generates a random 8-character hexadecimal ID.
func GenerateID() string {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}
