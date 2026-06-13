package rides

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/santosh/ride-app/pkg/db"
)

func RequestRide(ctx context.Context, riderID int64, pickupLat, pickupLng, dropoffLat, dropoffLng float64, pickupAddress, dropoffAddress, idempotencyKey string) (*Ride, error) {
	// Check if rider already has active ride
	existing, err := GetActiveRideByRiderID(ctx, riderID)
	if existing != nil {
		return nil, fmt.Errorf("rider already has an active ride: %d", existing.ID)
	}

	// Calculate fare
	fare, err := CalculateFare(ctx, pickupLat, pickupLng, dropoffLat, dropoffLng)
	if err != nil {
		return nil, fmt.Errorf("fare calculation failed: %v", err)
	}

	// Auto generate idempotency key if not provided
	if idempotencyKey == "" {
		idempotencyKey = uuid.New().String()
	}

	// Create ride
	ride := &Ride{
		RiderID:         riderID,
		PickupLat:       pickupLat,
		PickupLng:       pickupLng,
		PickupAddress:   pickupAddress,
		DropoffLat:      dropoffLat,
		DropoffLng:      dropoffLng,
		DropoffAddress:  dropoffAddress,
		FareMinPaisa:    fare.FareMinPaisa,
		FareMaxPaisa:    fare.FareMaxPaisa,
		DistanceMeters:  fare.DistanceMeters,
		DurationMinutes: fare.DurationMinutes,
		IdempotencyKey:  idempotencyKey,
	}

	rideID, err := CreateRide(ctx, ride)
	if err != nil {
		return nil, fmt.Errorf("failed to create ride: %v", err)
	}

	return GetRideByID(ctx, rideID)
}

func CancelRide(ctx context.Context, rideID int64, cancelledBy string, reason string) error {
	ride, err := GetRideByID(ctx, rideID)
	if err != nil {
		return fmt.Errorf("ride not found: %v", err)
	}

	if ride.Status == "completed" || ride.Status == "cancelled" {
		return fmt.Errorf("ride cannot be cancelled in status: %s", ride.Status)
	}

	_, err = db.Pool.Exec(ctx,
		`UPDATE rides SET status = 'cancelled', cancelled_by = $1,
		 cancel_reason = $2, updated_at = NOW()
		 WHERE id = $3`,
		cancelledBy, reason, rideID,
	)
	return err
}

func GetFareEstimate(ctx context.Context, pickupLat, pickupLng, dropoffLat, dropoffLng float64) (*FareEstimate, error) {
	return CalculateFare(ctx, pickupLat, pickupLng, dropoffLat, dropoffLng)
}
