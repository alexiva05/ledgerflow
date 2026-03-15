-- +goose Up
CREATE TYPE account_status AS ENUM ('active', 'frozen', 'closed');

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner UUID NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status account_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TYPE entry_direction AS ENUM ('debit', 'credit');

CREATE TABLE journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL,
    transaction_id UUID NOT NULL,
    direction entry_direction NOT NULL,
    amount NUMERIC(20, 4) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT fk_journal_account FOREIGN KEY (account_id) REFERENCES accounts(id) 
);

CREATE INDEX idx_journal_account_id ON journal_entries(account_id);
CREATE INDEX idx_journal_transaction_id ON journal_entries(transaction_id);

-- +goose Down
DROP TABLE journal_entries CASCADE;
DROP TABLE accounts CASCADE;
DROP TYPE account_status;
DROP TYPE entry_direction;