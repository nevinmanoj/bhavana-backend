package auth

import (
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// Generated passwords are emailed to non-technical users who type them on a
// phone: lowercase and digits only, without the lookalikes i l o 0 1.
// 20 characters from 31 symbols is about 99 bits.
const (
	generatedPasswordLength   = 20
	generatedPasswordAlphabet = "abcdefghjkmnpqrstuvwxyz23456789"
)

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
}

// GeneratePassword returns a random password for emailing. rand.Int per
// character avoids the modulo bias of indexing with a raw random byte.
func GeneratePassword() (string, error) {
	max := big.NewInt(int64(len(generatedPasswordAlphabet)))
	out := make([]byte, generatedPasswordLength)
	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = generatedPasswordAlphabet[n.Int64()]
	}
	return string(out), nil
}
