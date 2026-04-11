-- +goose Up

CREATE TYPE audit_event_type AS ENUM ('transaction.created', 'transaction.completed', 'transaction.failed', 'balance.updated', 'fraud.alert');

CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id TEXT NOT NULL,
    topic VARCHAR(255) NOT NULL,
    event_type audit_event_type NOT NULL,
    payload JSONB NOT NULL,
    hmac TEXT NOT NULL,
    prev_hmac TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_created_at ON audit_log (created_at);

-- +goose Down
DROP TABLE audit_log CASCADE;
DROP TYPE audit_event_type;