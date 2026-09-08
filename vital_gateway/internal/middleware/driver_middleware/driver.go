package driver_middleware

import (
	"bombe_main_server/internal/utils"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	Driver_id string `json:"driver_id"`
	jwt.RegisteredClaims
}

type contextKey string

const ClaimsContextKey contextKey = "Driver_id"

func DriverAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		access_token, err := r.Cookie("driver_access_token")

		if err != nil {
			utils.RespondWithError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var access_secret = os.Getenv("JWT_DRIVER_ACCESS_TOKEN_SECRET")

		claims, err := VerifyToken(access_token.Value, access_secret)

		if err != nil {
			utils.RespondWithError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims.Driver_id)

		next.ServeHTTP(w, r.WithContext(ctx))

	}

}

func VerifyToken(tokenString string, access_secret string) (*CustomClaims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(access_secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token has expired")
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}
