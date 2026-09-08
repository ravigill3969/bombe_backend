	package middleware
	
	import (
		"context"
		"errors"
		"fmt"
		"os"
		"strings"
	
		"github.com/golang-jwt/jwt/v5"
		"google.golang.org/grpc"
		"google.golang.org/grpc/codes"
		"google.golang.org/grpc/metadata"
		"google.golang.org/grpc/status"
	)
	
	var publicMethods = map[string]bool{
		"/auth.DriverAuthService/LoginDriver":    true,
		"/auth.DriverAuthService/RegisterDriver": true,
		"/auth.RiderAuthService/LoginRider":      true,
		"/auth.RiderAuthService/RegisterRider":   true,
		// "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo": true,
	}
	
	type contextKey string
	
	const Driver_id contextKey = "Driver_id"
	const Rider_id contextKey = "Rider_id"
	
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
	
	func GrpcAuthInterceptor(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
	
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}
	
		secret := os.Getenv("AUTH_GRPC_TOKEN_SECRET")
		if secret == "" {
			return nil, status.Error(codes.Internal, "Internal server error")
		}
	
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is missing")
		}
	
		authHeaders := md["authorization"]
		if len(authHeaders) == 0 || authHeaders[0] == "" {
			return nil, status.Error(codes.Unauthenticated, "authorization token is missing")
		}
	
		rawTokenHeader := authHeaders[0]
	
		if strings.HasPrefix(strings.ToLower(rawTokenHeader), "bearer ") {
			rawTokenHeader = rawTokenHeader[7:]
		}
		rawTokenHeader = strings.TrimSpace(rawTokenHeader)
	
		claims, err := VerifyToken(rawTokenHeader, secret)
	
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "Unauthorized")
		}
	
		if claims.Iss != "rest-api-gateway" || claims.Aud != "auth-grpc-service" {
			return nil, status.Error(codes.PermissionDenied, "Permission denied")
		}
	
		var newCtx context.Context
	
		if claims.Role == UserRole(RoleDriver) {
			newCtx = context.WithValue(ctx, Driver_id, claims.UserID)
		}
	
		if claims.Role == UserRole(RoleRider) {
			newCtx = context.WithValue(ctx, Rider_id, claims.UserID)
		}
	
		return handler(newCtx, req)
	}
	
	func VerifyToken(tokenString string, access_secret string) (*UserClaims, error) {
	
		token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
	
			if !ok {
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
	
		if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
			return claims, nil
		}
	
		return nil, errors.New("invalid token claims")
	}
