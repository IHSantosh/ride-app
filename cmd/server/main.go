package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/santosh/ride-app/internal/auth"
	"github.com/santosh/ride-app/internal/drivers"
	"github.com/santosh/ride-app/internal/rides"
	"github.com/santosh/ride-app/internal/users"
	"github.com/santosh/ride-app/internal/wallet"
	"github.com/santosh/ride-app/pkg/cache"
	"github.com/santosh/ride-app/pkg/db"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}

	db.Connect()
	defer db.Close()
	cache.Connect()
	defer cache.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "ride-app",
			"version": "0.1.0",
		})
	})

	// Auth routes
	mux.HandleFunc("/v1/auth/otp/send", auth.SendOTPHandler)
	mux.HandleFunc("/v1/auth/otp/verify", auth.VerifyOTPHandler)
	mux.HandleFunc("/v1/auth/refresh", auth.RefreshTokenHandler)

	// Rider routes (protected)
	mux.HandleFunc("/v1/rider/profile", auth.JWTMiddleware(users.GetRiderProfileHandler))
	mux.HandleFunc("/v1/rider/profile/update", auth.JWTMiddleware(users.UpdateRiderProfileHandler))

	// Driver routes (protected)
	mux.HandleFunc("/v1/driver/register", auth.JWTMiddleware(drivers.RegisterDriverHandler))
	mux.HandleFunc("/v1/driver/profile", auth.JWTMiddleware(drivers.GetDriverProfileHandler))

	// Wallet routes (protected)
	mux.HandleFunc("/v1/wallet", auth.JWTMiddleware(wallet.GetWalletHandler))
	mux.HandleFunc("/v1/wallet/topup", auth.JWTMiddleware(wallet.TopUpHandler))

	// Rides routes
	mux.HandleFunc("/v1/rides/request", auth.JWTMiddleware(rides.RequestRideHandler))
	mux.HandleFunc("/v1/rides/fare-estimate", rides.FareEstimateHandler)
	mux.HandleFunc("/v1/rides/", auth.JWTMiddleware(rides.GetRideHandler))
	mux.HandleFunc("/v1/rides/cancel/", auth.JWTMiddleware(rides.CancelRideHandler))

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
