CREATE TYPE ride_status AS ENUM (
    'requested',
    'searching',
    'matched',
    'driver_arrived',
    'in_progress',
    'completed',
    'cancelled'
);

CREATE TABLE rides (
    id BIGSERIAL PRIMARY KEY,
    rider_id BIGINT NOT NULL REFERENCES users(id),
    driver_id BIGINT REFERENCES users(id),
    status ride_status NOT NULL DEFAULT 'requested',
    pickup_lat DECIMAL(10,7) NOT NULL,
    pickup_lng DECIMAL(10,7) NOT NULL,
    pickup_address VARCHAR(255),
    dropoff_lat DECIMAL(10,7) NOT NULL,
    dropoff_lng DECIMAL(10,7) NOT NULL,
    dropoff_address VARCHAR(255),
    fare_min_paisa BIGINT,
    fare_max_paisa BIGINT,
    final_fare_paisa BIGINT,
    distance_meters INT,
    duration_minutes INT,
    idempotency_key VARCHAR(100) UNIQUE NOT NULL,
    cancelled_by VARCHAR(20),
    cancel_reason VARCHAR(255),
    requested_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    matched_at TIMESTAMPTZ,
    pickup_at TIMESTAMPTZ,
    dropoff_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rides_rider ON rides(rider_id);
CREATE INDEX idx_rides_driver ON rides(driver_id) WHERE driver_id IS NOT NULL;
CREATE INDEX idx_rides_status ON rides(status);
