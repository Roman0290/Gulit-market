CREATE TABLE addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label VARCHAR(64),
    line1 VARCHAR(255) NOT NULL,
    city VARCHAR(255) NOT NULL,
    lat DECIMAL(9, 6),
    lng DECIMAL(9, 6),
    is_default BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_addresses_user_id ON addresses(user_id);
