package rides

import (
	"context"
	"math"

	"github.com/santosh/ride-app/pkg/db"
)

type FareConfig struct {
	BaseFarePaisa    int64   `json:"base_fare_paisa"`
	PerKmPaisa       int64   `json:"per_km_paisa"`
	PerMinPaisa      int64   `json:"per_min_paisa"`
	ZoneMultiplier   float64 `json:"zone_multiplier"`
}

type FareEstimate struct {
	FareMinPaisa    int64 `json:"fare_min_paisa"`
	FareMaxPaisa    int64 `json:"fare_max_paisa"`
	DistanceMeters  int   `json:"distance_meters"`
	DurationMinutes int   `json:"duration_minutes"`
}

// Haversine formula — straight line distance between two coordinates
func haversineMeters(lat1, lng1, lat2, lng2 float64) int {
	const R = 6371000 // Earth radius in meters
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	dphi := (lat2 - lat1) * math.Pi / 180
	dlambda := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(dphi/2)*math.Sin(dphi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
			math.Sin(dlambda/2)*math.Sin(dlambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return int(R * c)
}

// Estimate duration based on distance (avg 30km/h in Kathmandu traffic)
func estimateDurationMinutes(distanceMeters int) int {
	avgSpeedMps := 30000.0 / 3600.0 // 30km/h in m/s
	return int(math.Ceil(float64(distanceMeters) / avgSpeedMps / 60))
}

func GetFareConfig(ctx context.Context) (*FareConfig, error) {
	config := &FareConfig{}
	err := db.Pool.QueryRow(ctx,
		`SELECT base_fare_paisa, per_km_paisa, per_min_paisa, zone_multiplier
		 FROM pricing_config
		 WHERE is_active = TRUE
		 LIMIT 1`,
	).Scan(
		&config.BaseFarePaisa,
		&config.PerKmPaisa,
		&config.PerMinPaisa,
		&config.ZoneMultiplier,
	)
	if err != nil {
		// Default Kathmandu pricing if no config found
		return &FareConfig{
			BaseFarePaisa:  10000, // NPR 100 base
			PerKmPaisa:     5000,  // NPR 50 per km
			PerMinPaisa:    100,   // NPR 1 per min
			ZoneMultiplier: 1.0,
		}, nil
	}
	return config, nil
}

func CalculateFare(ctx context.Context, pickupLat, pickupLng, dropoffLat, dropoffLng float64) (*FareEstimate, error) {
	distanceMeters := haversineMeters(pickupLat, pickupLng, dropoffLat, dropoffLng)
	durationMinutes := estimateDurationMinutes(distanceMeters)

	config, err := GetFareConfig(ctx)
	if err != nil {
		return nil, err
	}

	distanceKm := float64(distanceMeters) / 1000.0
	baseFare := float64(config.BaseFarePaisa)
	distanceFare := float64(config.PerKmPaisa) * distanceKm
	timeFare := float64(config.PerMinPaisa) * float64(durationMinutes)

	totalFare := (baseFare + distanceFare + timeFare) * config.ZoneMultiplier

	// Return min/max range (±10%)
	fareMin := int64(totalFare * 0.9)
	fareMax := int64(totalFare * 1.1)

	return &FareEstimate{
		FareMinPaisa:    fareMin,
		FareMaxPaisa:    fareMax,
		DistanceMeters:  distanceMeters,
		DurationMinutes: durationMinutes,
	}, nil
}
