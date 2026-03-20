-- +goose Up
CREATE TYPE transaction_status AS ENUM ('pending', 'completed', 'failed', 'cancelled');

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_account_id UUID NOT NULL,
    to_account_id UUID NOT NULL,
    amount NUMERIC(20, 4) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status transaction_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE TYPE idempotency_keys_status AS ENUM ('processing', 'completed');

CREATE TABLE idempotency_keys (
    key UUID PRIMARY KEY,
    response_status INTEGER,
    response_body TEXT,
    status idempotency_keys_status NOT NULL DEFAULT 'processing',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '24 hours'
);

CREATE TABLE outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic TEXT NOT NULL,
    key TEXT NOT NULL,
    payload JSONB NOT NULL,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outbox_sent_at ON outbox (sent_at) WHERE sent_at IS NULL;

-- +goose Down
DROP TABLE transactions CASCADE;
DROP TABLE idempotency_keys CASCADE;
DROP TABLE outbox CASCADE;
DROP TYPE transaction_status;
DROP TYPE idempotency_keys_status;