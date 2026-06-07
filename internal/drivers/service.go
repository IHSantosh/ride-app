package drivers

import (
	"context"
	"fmt"
)

func RegisterDriver(ctx context.Context, userID int64, licenseNumber, licenseExpiry string) (*DriverProfile, error) {
	// Check if driver profile already exists
	existing, _ := GetDriverProfileByUserID(ctx, userID)
	if existing != nil {
		return nil, fmt.Errorf("driver profile already exists")
	}

	// Create profile
	_, err := CreateDriverProfile(ctx, userID, licenseNumber, licenseExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to create driver profile: %v", err)
	}

	// Return created profile
	return GetDriverProfileByUserID(ctx, userID)
}
