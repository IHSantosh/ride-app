package drivers

import (
	"context"
	"time"

	"github.com/santosh/ride-app/pkg/db"
)

type DriverProfile struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	LicenseNumber string    `json:"license_number"`
	LicenseExpiry time.Time `json:"license_expiry"`
	TotalRides    int       `json:"total_rides"`
	TotalEarned   int64     `json:"total_earned_paisa"`
	AvgRating     float64   `json:"avg_rating"`
	IsOnline      bool      `json:"is_online"`
	IsBlocked     bool      `json:"is_blocked"`
	CurrentLat    float64   `json:"current_lat,omitempty"`
	CurrentLng    float64   `json:"current_lng,omitempty"`
}

func CreateDriverProfile(ctx context.Context, userID int64, licenseNumber, licenseExpiry string) (int64, error) {
	var id int64
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO driver_profiles (user_id, license_number, license_expiry)
		 VALUES ($1, $2, $3::date)
		 RETURNING id`,
		userID, licenseNumber, licenseExpiry,
	).Scan(&id)
	return id, err
}

func GetDriverProfileByUserID(ctx context.Context, userID int64) (*DriverProfile, error) {
	p := &DriverProfile{}
	var lat, lng *float64
	err := db.Pool.QueryRow(ctx,
		`SELECT id, user_id, license_number, license_expiry,
		        total_rides, total_earned_paisa, avg_rating,
		        is_online, is_blocked,
		        current_lat, current_lng
		 FROM driver_profiles
		 WHERE user_id = $1`,
		userID,
	).Scan(
		&p.ID, &p.UserID, &p.LicenseNumber, &p.LicenseExpiry,
		&p.TotalRides, &p.TotalEarned, &p.AvgRating,
		&p.IsOnline, &p.IsBlocked,
		&lat, &lng,
	)
	if err != nil {
		return nil, err
	}
	if lat != nil {
		p.CurrentLat = *lat
	}
	if lng != nil {
		p.CurrentLng = *lng
	}
	return p, nil
}
