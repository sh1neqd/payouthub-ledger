-- +goose Up
ALTER TABLE accounts
    ADD COLUMN kind TEXT NOT NULL DEFAULT 'wallet',
    ADD COLUMN currency TEXT NOT NULL DEFAULT 'RUB';

ALTER TABLE accounts
    DROP CONSTRAINT accounts_balance_nonnegative,
    DROP CONSTRAINT accounts_reserved_within_balance,

    ADD CONSTRAINT accounts_kind_valid
        CHECK (kind IN ('wallet', 'funding', 'settlement')),

    ADD CONSTRAINT accounts_currency_valid
        CHECK (currency = 'RUB'),

    ADD CONSTRAINT accounts_balance_valid
        CHECK (kind = 'funding' OR balance >= 0),

    ADD CONSTRAINT accounts_reserved_valid CHECK (
        (kind = 'wallet' AND reserved <= balance)
            OR
        (kind IN ('funding', 'settlement') AND reserved = 0)
        );

CREATE UNIQUE INDEX accounts_system_kind_currency_unique
    ON accounts (kind, currency)
    WHERE kind IN ('funding', 'settlement');

INSERT INTO accounts (kind, currency)
VALUES ('funding', 'RUB'), ('settlement', 'RUB');

-- +goose Down
-- +goose StatementBegin
DO $$
    BEGIN
        IF EXISTS (
            SELECT 1
            FROM accounts
            WHERE kind IN ('funding', 'settlement')
              AND (balance <> 0 OR reserved <> 0)
        ) THEN
            RAISE EXCEPTION 'cannot rollback: system accounts are not empty';
        END IF;
    END;
$$;
-- +goose StatementEnd

DELETE FROM accounts
WHERE kind IN ('funding', 'settlement');

DROP INDEX accounts_system_kind_currency_unique;

ALTER TABLE accounts
    DROP CONSTRAINT accounts_kind_valid,
    DROP CONSTRAINT accounts_currency_valid,
    DROP CONSTRAINT accounts_balance_valid,
    DROP CONSTRAINT accounts_reserved_valid,
    DROP COLUMN kind,
    DROP COLUMN currency,

    ADD CONSTRAINT accounts_balance_nonnegative
        CHECK (balance >= 0),

    ADD CONSTRAINT accounts_reserved_within_balance
        CHECK (reserved <= balance);