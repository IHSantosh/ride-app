package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/santosh/ride-app/pkg/db"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}

	db.Connect()
	defer db.Close()

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
			"version": "0.0.4",
		})
	})

	// Test: insert a user and read back
	mux.HandleFunc("/test/user", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		// Insert test user
		_, err := db.Pool.Exec(ctx,
			`INSERT INTO users (phone_number, country_code, role, status, full_name)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (phone_number) DO NOTHING`,
			"+9779800000001", "+977", 1, 1, "Test Rider",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Read back
		var id int64
		var name, phone string
		var role int
		err = db.Pool.QueryRow(ctx,
			`SELECT id, full_name, phone_number, role FROM users WHERE phone_number = $1`,
			"+9779800000001",
		).Scan(&id, &name, &phone, &role)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    id,
			"name":  name,
			"phone": phone,
			"role":  role,
		})
	})

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
