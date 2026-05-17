package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/santosh/ride-app/pkg/cache"
)

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "refresh_token is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Get userID from Redis
	key := fmt.Sprintf("refresh:%s", req.RefreshToken)
	val, err := cache.Client.Get(ctx, key).Result()
	if err != nil {
		http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		http.Error(w, "invalid token data", http.StatusInternalServerError)
		return
	}

	// Generate new access token (default role 1, improve later)
	accessToken, err := GenerateAccessToken(userID, 1)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": accessToken,
		"expires_in":   15 * 60,
	})
}
