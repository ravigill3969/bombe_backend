CREATE TYPE service_type AS ENUM ('unspecified', 'economy', 'confort', 'XL');

CREATE TABLE payment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    temp_ride_id UUID NOT NULL,
    stripe_session_id VARCHAR(255),
    --
    rider_id UUID NOT NULL,
    service_type service_type DEFAULT 'unspecified',
    pickup_latitude DECIMAL(10, 8) NOT NULL,
    pickup_longitude DECIMAL(11, 8),
    pickup_address TEXT NOT NULL,
    --
    dropoff_latitude DECIMAL(10, 8),
    dropoff_longitude DECIMAL(11, 8),
    dropoff_address TEXT,
    fare_amount_in_cents BIGINT NOT NULL CHECK (fare_amount_in_cents >= 0),
    fare_currency VARCHAR(3) DEFAULT 'CAD' NOT NULL,
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);