package drivers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/santosh/ride-app/internal/auth"
)

type RegisterDriverRequest struct {
	LicenseNumber string `json:"license_number"`
	LicenseExpiry string `json:"license_expiry"`
}

func RegisterDriverHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(auth.UserIDKey).(int64)

	var req RegisterDriverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.LicenseNumber == "" || req.LicenseExpiry == "" {
		http.Error(w, "license_number and license_expiry are required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	profile, err := RegisterDriver(ctx, userID, req.LicenseNumber, req.LicenseExpiry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func GetDriverProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(auth.UserIDKey).(int64)
	ctx := context.Background()

	profile, err := GetDriverProfileByUserID(ctx, userID)
	if err != nil {
		http.Error(w, "driver profile not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}
