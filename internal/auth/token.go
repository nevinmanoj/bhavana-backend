package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// NewResetToken returns the token that goes in the reset link and the hash
// that is stored. Only the hash is ever written to the database.
func NewResetToken() (token string, hash []byte, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, err
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	return token, HashResetToken(token), nil
}

func HashResetToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
