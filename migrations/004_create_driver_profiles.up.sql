CREATE TABLE driver_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    license_number VARCHAR(50) UNIQUE NOT NULL,
    license_expiry DATE NOT NULL,
    total_rides INT DEFAULT 0,
    total_earned_paisa BIGINT DEFAULT 0,
    avg_rating DECIMAL(3,2) DEFAULT 5.00,
    is_online BOOLEAN DEFAULT FALSE,
    is_blocked BOOLEAN DEFAULT FALSE,
    block_reason VARCHAR(255),
    current_lat DECIMAL(10,7),
    current_lng DECIMAL(10,7),
    last_location_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_driver_profiles_user ON driver_profiles(user_id);
CREATE INDEX idx_driver_online ON driver_profiles(is_online) WHERE is_blocked = FALSE;
