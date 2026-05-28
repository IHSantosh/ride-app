CREATE TABLE rider_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    emergency_contact_name VARCHAR(100),
    emergency_contact_phone VARCHAR(20),
    emergency_contact_relation VARCHAR(50),
    preferred_language VARCHAR(10) DEFAULT 'en',
    low_data_mode BOOLEAN DEFAULT FALSE,
    notifications_enabled BOOLEAN DEFAULT TRUE,
    total_rides INT DEFAULT 0,
    total_spent_paisa BIGINT DEFAULT 0,
    avg_rating DECIMAL(3,2) DEFAULT 5.00,
    is_blocked BOOLEAN DEFAULT FALSE,
    block_reason VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rider_profiles_user ON rider_profiles(user_id);
