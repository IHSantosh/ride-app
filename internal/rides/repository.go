package rides

import (
	"context"
	"time"

	"github.com/santosh/ride-app/pkg/db"
)

type Ride struct {
	ID              int64      `json:"id"`
	RiderID         int64      `json:"rider_id"`
	DriverID        *int64     `json:"driver_id,omitempty"`
	Status          string     `json:"status"`
	PickupLat       float64    `json:"pickup_lat"`
	PickupLng       float64    `json:"pickup_lng"`
	PickupAddress   string     `json:"pickup_address,omitempty"`
	DropoffLat      float64    `json:"dropoff_lat"`
	DropoffLng      float64    `json:"dropoff_lng"`
	DropoffAddress  string     `json:"dropoff_address,omitempty"`
	FareMinPaisa    int64      `json:"fare_min_paisa"`
	FareMaxPaisa    int64      `json:"fare_max_paisa"`
	FinalFarePaisa  *int64     `json:"final_fare_paisa,omitempty"`
	DistanceMeters  int        `json:"distance_meters"`
	DurationMinutes int        `json:"duration_minutes"`
	IdempotencyKey  string     `json:"idempotency_key"`
	RequestedAt     time.Time  `json:"requested_at"`
	MatchedAt       *time.Time `json:"matched_at,omitempty"`
	PickupAt        *time.Time `json:"pickup_at,omitempty"`
	DropoffAt       *time.Time `json:"dropoff_at,omitempty"`
}

func CreateRide(ctx context.Context, r *Ride) (int64, error) {
	var id int64
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO rides (
			rider_id, status, pickup_lat, pickup_lng, pickup_address,
			dropoff_lat, dropoff_lng, dropoff_address,
			fare_min_paisa, fare_max_paisa,
			distance_meters, duration_minutes, idempotency_key
		) VALUES ($1,'requested',$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (idempotency_key) DO UPDATE SET updated_at = NOW()
		RETURNING id`,
		r.RiderID, r.PickupLat, r.PickupLng, r.PickupAddress,
		r.DropoffLat, r.DropoffLng, r.DropoffAddress,
		r.FareMinPaisa, r.FareMaxPaisa,
		r.DistanceMeters, r.DurationMinutes, r.IdempotencyKey,
	).Scan(&id)
	return id, err
}

func GetRideByID(ctx context.Context, rideID int64) (*Ride, error) {
	r := &Ride{}
	err := db.Pool.QueryRow(ctx,
		`SELECT id, rider_id, driver_id, status,
			pickup_lat, pickup_lng, COALESCE(pickup_address,''),
			dropoff_lat, dropoff_lng, COALESCE(dropoff_address,''),
			fare_min_paisa, fare_max_paisa, final_fare_paisa,
			distance_meters, duration_minutes, idempotency_key,
			requested_at, matched_at, pickup_at, dropoff_at
		 FROM rides WHERE id = $1`,
		rideID,
	).Scan(
		&r.ID, &r.RiderID, &r.DriverID, &r.Status,
		&r.PickupLat, &r.PickupLng, &r.PickupAddress,
		&r.DropoffLat, &r.DropoffLng, &r.DropoffAddress,
		&r.FareMinPaisa, &r.FareMaxPaisa, &r.FinalFarePaisa,
		&r.DistanceMeters, &r.DurationMinutes, &r.IdempotencyKey,
		&r.RequestedAt, &r.MatchedAt, &r.PickupAt, &r.DropoffAt,
	)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func UpdateRideStatus(ctx context.Context, rideID int64, status string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE rides SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, rideID,
	)
	return err
}

func GetActiveRideByRiderID(ctx context.Context, riderID int64) (*Ride, error) {
	r := &Ride{}
	err := db.Pool.QueryRow(ctx,
		`SELECT id, rider_id, driver_id, status,
			pickup_lat, pickup_lng, COALESCE(pickup_address,''),
			dropoff_lat, dropoff_lng, COALESCE(dropoff_address,''),
			fare_min_paisa, fare_max_paisa, final_fare_paisa,
			distance_meters, duration_minutes, idempotency_key,
			requested_at, matched_at, pickup_at, dropoff_at
		 FROM rides
		 WHERE rider_id = $1
		 AND status NOT IN ('completed','cancelled')
		 ORDER BY created_at DESC LIMIT 1`,
		riderID,
	).Scan(
		&r.ID, &r.RiderID, &r.DriverID, &r.Status,
		&r.PickupLat, &r.PickupLng, &r.PickupAddress,
		&r.DropoffLat, &r.DropoffLng, &r.DropoffAddress,
		&r.FareMinPaisa, &r.FareMaxPaisa, &r.FinalFarePaisa,
		&r.DistanceMeters, &r.DurationMinutes, &r.IdempotencyKey,
		&r.RequestedAt, &r.MatchedAt, &r.PickupAt, &r.DropoffAt,
	)
	if err != nil {
		return nil, err
	}
	return r, nil
}
