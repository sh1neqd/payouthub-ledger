-- +goose Up
CREATE TABLE holds (
                       id UUID PRIMARY KEY,
                       payout_id UUID NOT NULL UNIQUE,
                       account_id BIGINT NOT NULL REFERENCES accounts(id),
                       amount BIGINT NOT NULL,
                       currency TEXT NOT NULL,
                       status TEXT NOT NULL DEFAULT 'active',

                       reserve_operation_id UUID NOT NULL UNIQUE REFERENCES operations(id),
                       capture_operation_id UUID UNIQUE REFERENCES operations(id),
                       release_operation_id UUID UNIQUE REFERENCES operations(id),

                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       completed_at TIMESTAMPTZ,

                       CONSTRAINT holds_amount_positive
                           CHECK (amount > 0),

                       CONSTRAINT holds_currency_valid
                           CHECK (currency = 'RUB'),

                       CONSTRAINT holds_state_valid CHECK (
                           (
                               status = 'active'
                                   AND capture_operation_id IS NULL
                                   AND release_operation_id IS NULL
                                   AND completed_at IS NULL
                               )
                               OR
                           (
                               status = 'captured'
                                   AND capture_operation_id IS NOT NULL
                                   AND release_operation_id IS NULL
                                   AND completed_at IS NOT NULL
                               )
                               OR
                           (
                               status = 'released'
                                   AND capture_operation_id IS NULL
                                   AND release_operation_id IS NOT NULL
                                   AND completed_at IS NOT NULL
                               )
                           )
);

CREATE INDEX holds_account_status_idx
    ON holds (account_id, status);

-- +goose Down
DROP TABLE holds;