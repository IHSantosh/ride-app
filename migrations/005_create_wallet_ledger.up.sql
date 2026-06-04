CREATE TYPE ledger_type AS ENUM ('topup', 'ride_payment', 'refund', 'cancellation_fee');

CREATE TABLE wallet_ledger (
    id BIGSERIAL PRIMARY KEY,
    wallet_id BIGINT NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    type ledger_type NOT NULL,
    amount_paisa BIGINT NOT NULL,
    ride_id BIGINT,
    ref_id VARCHAR(100),
    idempotency_key VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_ledger_wallet ON wallet_ledger(wallet_id);
CREATE INDEX idx_ledger_ride ON wallet_ledger(ride_id) WHERE ride_id IS NOT NULL;
