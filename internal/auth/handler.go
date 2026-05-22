package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/santosh/ride-app/pkg/db"
)

type SendOTPRequest struct {
	Phone       string `json:"phone"`
	CountryCode string `json:"country_code"`
}

type VerifyOTPRequest struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
	Role  int    `json:"role"`
}

func SendOTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SendOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Phone == "" {
		http.Error(w, "phone is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	otp, err := StoreOTP(ctx, req.Phone)
	if err != nil {
		http.Error(w, err.Error(), http.StatusTooManyRequests)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if os.Getenv("ENV") == "development" {
		json.NewEncoder(w).Encode(map[string]string{
			"message": "OTP sent successfully",
			"otp":     otp,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "OTP sent successfully",
	})
}

func VerifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Phone == "" || req.OTP == "" {
		http.Error(w, "phone and otp are required", http.StatusBadRequest)
		return
	}

	if req.Role == 0 {
		req.Role = 1
	}

	ctx := context.Background()

	valid, err := VerifyOTP(ctx, req.Phone, req.OTP)
	if err != nil || !valid {
		http.Error(w, "invalid or expired OTP", http.StatusUnauthorized)
		return
	}

	var userID int64
	err = db.Pool.QueryRow(ctx,
		`SELECT id FROM users WHERE phone_number = $1 AND deleted_at IS NULL`,
		req.Phone,
	).Scan(&userID)

	if err != nil {
		err = db.Pool.QueryRow(ctx,
			`INSERT INTO users (phone_number, country_code, role, status, full_name)
			 VALUES ($1, '+977', $2, 1, 'New User')
			 RETURNING id`,
			req.Phone, req.Role,
		).Scan(&userID)

		if err != nil {
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			return
		}

		_, err = db.Pool.Exec(ctx,
			`INSERT INTO wallets (user_id, balance_paisa) VALUES ($1, 0)`,
			userID,
		)
		if err != nil {
			http.Error(w, "failed to create wallet", http.StatusInternalServerError)
			return
		}
	}

	accessToken, err := GenerateAccessToken(userID, req.Role)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := GenerateRefreshToken(ctx, userID)
	if err != nil {
		http.Error(w, "failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user_id":       userID,
		"role":          req.Role,
	})
}
