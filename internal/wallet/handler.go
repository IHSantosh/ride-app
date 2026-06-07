package wallet

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/santosh/ride-app/internal/auth"
	"github.com/google/uuid"
)

func GetWalletHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(auth.UserIDKey).(int64)
	ctx := context.Background()

	wallet, transactions, err := GetWalletWithTransactions(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"wallet":       wallet,
		"transactions": transactions,
	})
}

func TopUpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(auth.UserIDKey).(int64)

	var req struct {
		AmountPaisa    int64  `json:"amount_paisa"`
		IdempotencyKey string `json:"idempotency_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.AmountPaisa <= 0 {
		http.Error(w, "amount_paisa must be greater than 0", http.StatusBadRequest)
		return
	}

	// Auto generate idempotency key if not provided
	if req.IdempotencyKey == "" {
		req.IdempotencyKey = uuid.New().String()
	}

	ctx := context.Background()
	wallet, err := TopUpWallet(ctx, userID, req.AmountPaisa, req.IdempotencyKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "topup successful",
		"wallet":  wallet,
	})
}
