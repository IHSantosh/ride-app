CREATE TABLE wallets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    balance_paisa BIGINT DEFAULT 0,
    max_balance_paisa BIGINT DEFAULT 1000000,
    min_topup_paisa BIGINT DEFAULT 5000,
    is_frozen BOOLEAN DEFAULT FALSE,
    frozen_reason VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_wallets_user ON wallets(user_id);
