package users

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/santosh/ride-app/internal/auth"
	"github.com/santosh/ride-app/pkg/db"
)

type RiderProfile struct {
	ID                       int64     `json:"id"`
	UserID                   int64     `json:"user_id"`
	FullName                 string    `json:"full_name"`
	PhoneNumber              string    `json:"phone_number"`
	EmergencyContactName     *string   `json:"emergency_contact_name"`
	EmergencyContactPhone    *string   `json:"emergency_contact_phone"`
	EmergencyContactRelation *string   `json:"emergency_contact_relation"`
	PreferredLanguage        string    `json:"preferred_language"`
	LowDataMode              bool      `json:"low_data_mode"`
	TotalRides               int       `json:"total_rides"`
	AvgRating                float64   `json:"avg_rating"`
	CreatedAt                time.Time `json:"created_at"`
}

type UpdateRiderRequest struct {
	FullName                 string `json:"full_name"`
	EmergencyContactName     string `json:"emergency_contact_name"`
	EmergencyContactPhone    string `json:"emergency_contact_phone"`
	EmergencyContactRelation string `json:"emergency_contact_relation"`
	PreferredLanguage        string `json:"preferred_language"`
	LowDataMode              bool   `json:"low_data_mode"`
}

func GetRiderProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(auth.UserIDKey).(int64)
	ctx := context.Background()

	var profile RiderProfile
	err := db.Pool.QueryRow(ctx, `
		SELECT 
			rp.id, rp.user_id, u.full_name, u.phone_number,
			rp.emergency_contact_name, rp.emergency_contact_phone,
			rp.emergency_contact_relation, rp.preferred_language,
			rp.low_data_mode, rp.total_rides, rp.avg_rating, rp.created_at
		FROM rider_profiles rp
		JOIN users u ON u.id = rp.user_id
		WHERE rp.user_id = $1
	`, userID).Scan(
		&profile.ID, &profile.UserID, &profile.FullName, &profile.PhoneNumber,
		&profile.EmergencyContactName, &profile.EmergencyContactPhone,
		&profile.EmergencyContactRelation, &profile.PreferredLanguage,
		&profile.LowDataMode, &profile.TotalRides, &profile.AvgRating,
		&profile.CreatedAt,
	)

	if err != nil {
		_, err = db.Pool.Exec(ctx,
			`INSERT INTO rider_profiles (user_id) VALUES ($1) ON CONFLICT DO NOTHING`,
			userID,
		)
		if err != nil {
			http.Error(w, "failed to create profile", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id": userID,
			"message": "profile created, please update your details",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func UpdateRiderProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(auth.UserIDKey).(int64)

	var req UpdateRiderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	if req.FullName != "" {
		_, err := db.Pool.Exec(ctx,
			`UPDATE users SET full_name = $1, updated_at = NOW() WHERE id = $2`,
			req.FullName, userID,
		)
		if err != nil {
			http.Error(w, "failed to update name", http.StatusInternalServerError)
			return
		}
	}

	_, err := db.Pool.Exec(ctx, `
		INSERT INTO rider_profiles (
			user_id, emergency_contact_name, emergency_contact_phone,
			emergency_contact_relation, preferred_language, low_data_mode
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET
			emergency_contact_name = EXCLUDED.emergency_contact_name,
			emergency_contact_phone = EXCLUDED.emergency_contact_phone,
			emergency_contact_relation = EXCLUDED.emergency_contact_relation,
			preferred_language = EXCLUDED.preferred_language,
			low_data_mode = EXCLUDED.low_data_mode,
			updated_at = NOW()
	`,
		userID,
		req.EmergencyContactName, req.EmergencyContactPhone,
		req.EmergencyContactRelation, req.PreferredLanguage,
		req.LowDataMode,
	)
	if err != nil {
		http.Error(w, "failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "profile updated successfully",
	})
}
