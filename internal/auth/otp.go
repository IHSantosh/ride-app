package auth

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/santosh/ride-app/pkg/cache"
)

const (
	otpExpiry     = 5 * time.Minute
	otpRateExpiry = 1 * time.Hour
	maxOTPPerHour = 3
)

func GenerateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func StoreOTP(ctx context.Context, phone string) (string, error) {
	// Check rate limit first
	rateKey := fmt.Sprintf("otp_rate:%s", phone)
	count, err := cache.Client.Incr(ctx, rateKey).Result()
	if err != nil {
		return "", fmt.Errorf("rate limit check failed: %v", err)
	}

	// Set expiry on first request
	if count == 1 {
		cache.Client.Expire(ctx, rateKey, otpRateExpiry)
	}

	// Block if too many requests
	if count > maxOTPPerHour {
		return "", fmt.Errorf("too many OTP requests, try again later")
	}

	// Generate and store OTP
	otp := GenerateOTP()
	otpKey := fmt.Sprintf("otp:%s", phone)
	err = cache.Client.Set(ctx, otpKey, otp, otpExpiry).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store OTP: %v", err)
	}

	return otp, nil
}

func VerifyOTP(ctx context.Context, phone, otp string) (bool, error) {
	otpKey := fmt.Sprintf("otp:%s", phone)

	// Get stored OTP
	stored, err := cache.Client.Get(ctx, otpKey).Result()
	if err == redis.Nil {
		return false, fmt.Errorf("OTP expired or not found")
	}
	if err != nil {
		return false, fmt.Errorf("failed to get OTP: %v", err)
	}

	// Compare
	if stored != otp {
		return false, fmt.Errorf("invalid OTP")
	}

	// Delete after successful verify (one time use)
	cache.Client.Del(ctx, otpKey)

	return true, nil
}
