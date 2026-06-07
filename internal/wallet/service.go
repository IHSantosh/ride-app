package wallet

import (
	"context"
	"fmt"
)

func GetWalletWithTransactions(ctx context.Context, userID int64) (*Wallet, []LedgerEntry, error) {
	wallet, err := GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("wallet not found: %v", err)
	}

	transactions, err := GetRecentTransactions(ctx, wallet.ID, 10)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get transactions: %v", err)
	}

	return wallet, transactions, nil
}

func TopUpWallet(ctx context.Context, userID int64, amountPaisa int64, idempotencyKey string) (*Wallet, error) {
	wallet, err := GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %v", err)
	}

	// Check frozen
	if wallet.IsFrozen {
		return nil, fmt.Errorf("wallet is frozen: %s", wallet.FrozenReason)
	}

	// Check min topup
	if amountPaisa < wallet.MinTopupPaisa {
		return nil, fmt.Errorf("minimum topup is %d paisa", wallet.MinTopupPaisa)
	}

	// Check max balance
	if wallet.BalancePaisa+amountPaisa > wallet.MaxBalancePaisa {
		return nil, fmt.Errorf("exceeds maximum wallet balance of %d paisa", wallet.MaxBalancePaisa)
	}

	// Add ledger entry
	err = AddLedgerEntry(ctx, wallet.ID, "topup", amountPaisa, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("topup failed: %v", err)
	}

	// Return updated wallet
	return GetWalletByUserID(ctx, userID)
}
