package wallet

import (
	"context"
	"fmt"

	"github.com/santosh/ride-app/pkg/db"
)

type Wallet struct {
	ID              int64  `json:"id"`
	UserID          int64  `json:"user_id"`
	BalancePaisa    int64  `json:"balance_paisa"`
	MaxBalancePaisa int64  `json:"max_balance_paisa"`
	MinTopupPaisa   int64  `json:"min_topup_paisa"`
	IsFrozen        bool   `json:"is_frozen"`
	FrozenReason    string `json:"frozen_reason,omitempty"`
}

type LedgerEntry struct {
	ID             int64  `json:"id"`
	WalletID       int64  `json:"wallet_id"`
	Type           string `json:"type"`
	AmountPaisa    int64  `json:"amount_paisa"`
	IdempotencyKey string `json:"idempotency_key"`
	CreatedAt      string `json:"created_at"`
}

func GetWalletByUserID(ctx context.Context, userID int64) (*Wallet, error) {
	w := &Wallet{}
	err := db.Pool.QueryRow(ctx,
		`SELECT id, user_id, balance_paisa, max_balance_paisa,
		        min_topup_paisa, is_frozen, COALESCE(frozen_reason, '')
		 FROM wallets WHERE user_id = $1`,
		userID,
	).Scan(
		&w.ID, &w.UserID, &w.BalancePaisa, &w.MaxBalancePaisa,
		&w.MinTopupPaisa, &w.IsFrozen, &w.FrozenReason,
	)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func AddLedgerEntry(ctx context.Context, walletID int64, entryType string, amountPaisa int64, idempotencyKey string) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO wallet_ledger (wallet_id, type, amount_paisa, idempotency_key)
		 VALUES ($1, $2, $3, $4)`,
		walletID, entryType, amountPaisa, idempotencyKey,
	)
	if err != nil {
		return fmt.Errorf("ledger entry failed: %v", err)
	}

	// Update wallet balance
	_, err = db.Pool.Exec(ctx,
		`UPDATE wallets SET balance_paisa = balance_paisa + $1, updated_at = NOW()
		 WHERE id = $2`,
		amountPaisa, walletID,
	)
	return err
}

func GetRecentTransactions(ctx context.Context, walletID int64, limit int) ([]LedgerEntry, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, wallet_id, type, amount_paisa, idempotency_key, created_at
		 FROM wallet_ledger
		 WHERE wallet_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`,
		walletID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LedgerEntry
	for rows.Next() {
		var e LedgerEntry
		if err := rows.Scan(&e.ID, &e.WalletID, &e.Type, &e.AmountPaisa, &e.IdempotencyKey, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
