CREATE TABLE platform_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    commission_rate DECIMAL(5, 2) NOT NULL DEFAULT 10.00 CHECK (commission_rate >= 0 AND commission_rate <= 100),
    tax_rate DECIMAL(5, 2) NOT NULL DEFAULT 0.00 CHECK (tax_rate >= 0 AND tax_rate <= 100),
    default_delivery_fee DECIMAL(10, 2) NOT NULL DEFAULT 0.00 CHECK (default_delivery_fee >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO platform_settings (id) VALUES (1);
