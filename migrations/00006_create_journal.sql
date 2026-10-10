-- +goose Up
CREATE TABLE journal_transactions (
                                      id UUID PRIMARY KEY,
                                      operation_id UUID NOT NULL UNIQUE REFERENCES operations(id),
                                      kind TEXT NOT NULL,
                                      currency TEXT NOT NULL,
                                      hold_id UUID UNIQUE REFERENCES holds(id),
                                      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                                      CONSTRAINT journal_transactions_currency_valid
                                          CHECK (currency = 'RUB'),

                                      CONSTRAINT journal_transactions_kind_hold_valid CHECK (
                                          (kind = 'fund' AND hold_id IS NULL)
                                              OR
                                          (kind = 'capture' AND hold_id IS NOT NULL)
                                          )
);

CREATE TABLE journal_entries (
                                 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                                 transaction_id UUID NOT NULL REFERENCES journal_transactions(id),
                                 account_id BIGINT NOT NULL REFERENCES accounts(id),
                                 delta BIGINT NOT NULL,

                                 CONSTRAINT journal_entries_delta_nonzero
                                     CHECK (delta <> 0),

                                 CONSTRAINT journal_entries_transaction_account_unique
                                     UNIQUE (transaction_id, account_id)
);

CREATE INDEX journal_entries_account_id_idx
    ON journal_entries (account_id, id);

-- +goose Down
DROP TABLE journal_entries;
DROP TABLE journal_transactions;