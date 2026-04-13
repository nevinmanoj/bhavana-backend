package auth

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

type Claims struct {
	UserID int64         `json:"user_id"`
	Email  string        `json:"email"`
	Role   rbac.UserRole `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(role rbac.UserRole, userID int64, email string, secret []byte) (string, error) {

	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ParseToken(tokenStr string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(
					"unexpected signing method: %v",
					token.Header["alg"],
				)
			}
			return secret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
