package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserRole string

const (
	RoleDriver UserRole = "driver"
	RoleRider  UserRole = "rider"
)

type UserClaims struct {
	UserID string   `json:"user_id"`
	Role   UserRole `json:"role"`
	Iss    string   `json:"iss"`
	Aud    string   `json:"aud"`
	jwt.RegisteredClaims
}

//aud is receiver server
func CreateToken(user_id string, role UserRole, expiresIn time.Duration, aud string) (string, error) { 

	secret := os.Getenv("AUTH_GRPC_TOKEN_SECRET")

	if secret == "" {
		return "", fmt.Errorf("Internal server error")
	}

	claims := UserClaims{
		UserID: user_id,
		Role:   role,
		Iss:    "rest-api-gateway",
		Aud:    aud,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "rest-api-gateway",
			Audience:  jwt.ClaimStrings{"auth-grpc-service"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}
