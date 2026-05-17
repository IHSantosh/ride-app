package auth

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/santosh/ride-app/pkg/cache"
)

type Claims struct {
	UserID int64 `json:"sub"`
	Role   int   `json:"r"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID int64, role int) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	expiryMinutes, _ := strconv.Atoi(os.Getenv("JWT_EXPIRY_MINUTES"))
	if expiryMinutes == 0 {
		expiryMinutes = 15
	}

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiryMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(ctx context.Context, userID int64) (string, error) {
	refreshToken := uuid.New().String()
	expiryDays, _ := strconv.Atoi(os.Getenv("REFRESH_EXPIRY_DAYS"))
	if expiryDays == 0 {
		expiryDays = 30
	}

	key := fmt.Sprintf("refresh:%s", refreshToken)
	err := cache.Client.Set(ctx, key, userID, time.Duration(expiryDays)*24*time.Hour).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store refresh token: %v", err)
	}

	return refreshToken, nil
}

func ValidateAccessToken(tokenString string) (*Claims, error) {
	secret := os.Getenv("JWT_SECRET")

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
