ALTER TABLE orders ALTER COLUMN payment_status DROP DEFAULT;
ALTER TABLE orders ALTER COLUMN payment_status TYPE VARCHAR(20);
DROP TYPE payment_status;
CREATE TYPE payment_status AS ENUM ('pending', 'paid', 'failed');
ALTER TABLE orders ALTER COLUMN payment_status TYPE payment_status USING payment_status::payment_status;
ALTER TABLE orders ALTER COLUMN payment_status SET DEFAULT 'pending';
