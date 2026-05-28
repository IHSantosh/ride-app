package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/santosh/ride-app/internal/auth"
	"github.com/santosh/ride-app/internal/users"
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
			"version": "0.0.7",
		})
	})

	// Auth routes
	mux.HandleFunc("/v1/auth/otp/send", auth.SendOTPHandler)
	mux.HandleFunc("/v1/auth/otp/verify", auth.VerifyOTPHandler)
	mux.HandleFunc("/v1/auth/refresh", auth.RefreshTokenHandler)

	// Rider routes (protected)
	mux.HandleFunc("/v1/rider/profile", auth.JWTMiddleware(users.GetRiderProfileHandler))
	mux.HandleFunc("/v1/rider/profile/update", auth.JWTMiddleware(users.UpdateRiderProfileHandler))

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
