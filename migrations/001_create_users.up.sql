CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    phone_number VARCHAR(20) UNIQUE NOT NULL,
    country_code VARCHAR(5) NOT NULL DEFAULT '+977',
    role SMALLINT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 1,
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(100),
    password_hash VARCHAR(255),
    is_verified BOOLEAN DEFAULT FALSE,
    verification_level SMALLINT DEFAULT 0,
    device_fingerprint VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT user_role_valid CHECK (role >= 1 AND role <= 5),
    CONSTRAINT user_status_valid CHECK (status >= 1 AND status <= 3)
);

CREATE INDEX idx_users_phone ON users(phone_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role_status ON users(role, status) WHERE deleted_at IS NULL;
