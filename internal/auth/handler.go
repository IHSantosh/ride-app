package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
)

type SendOTPRequest struct {
	Phone       string `json:"phone"`
	CountryCode string `json:"country_code"`
}

type VerifyOTPRequest struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
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

	// Dev mode — return OTP in response
	// Production — send via SMS, never return in response
	if os.Getenv("ENV") == "development" {
		json.NewEncoder(w).Encode(map[string]string{
			"message": "OTP sent successfully",
			"otp":     otp, // remove this in production
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

	ctx := context.Background()
	valid, err := VerifyOTP(ctx, req.Phone, req.OTP)
	if err != nil || !valid {
		http.Error(w, "invalid or expired OTP", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "OTP verified successfully",
		"phone":   req.Phone,
	})
}
