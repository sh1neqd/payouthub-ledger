-- +goose Up
CREATE TABLE accounts (
                          id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                          balance BIGINT NOT NULL DEFAULT 0,
                          reserved BIGINT NOT NULL DEFAULT 0,
                          created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                          CONSTRAINT accounts_balance_nonnegative
                              CHECK (balance >= 0),

                          CONSTRAINT accounts_reserved_nonnegative
                              CHECK (reserved >= 0),

                          CONSTRAINT accounts_reserved_within_balance
                              CHECK (reserved <= balance)
);

-- +goose Down
DROP TABLE accounts;