CREATE TYPE coupon_discount_type AS ENUM ('percent', 'fixed');

CREATE TABLE coupons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(32) NOT NULL UNIQUE,
    discount_type coupon_discount_type NOT NULL,
    discount_value DECIMAL(10, 2) NOT NULL CHECK (discount_value > 0),
    CONSTRAINT percent_discount_within_100 CHECK (discount_type != 'percent' OR discount_value <= 100),
    is_active BOOLEAN NOT NULL DEFAULT true,
    expires_at TIMESTAMPTZ,
    usage_limit INT CHECK (usage_limit IS NULL OR usage_limit > 0),
    times_used INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
