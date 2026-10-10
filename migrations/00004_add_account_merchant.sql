-- +goose Up
ALTER TABLE accounts
    ADD COLUMN merchant_id UUID;

UPDATE accounts
SET merchant_id = gen_random_uuid()
WHERE kind = 'wallet';

ALTER TABLE accounts
    ADD CONSTRAINT accounts_merchant_valid CHECK (
        (
            kind = 'wallet'
                AND merchant_id IS NOT NULL
                AND merchant_id <> '00000000-0000-0000-0000-000000000000'::UUID
            )
            OR
        (kind IN ('funding', 'settlement') AND merchant_id IS NULL)
        ),

    ADD CONSTRAINT accounts_merchant_currency_unique
        UNIQUE (merchant_id, currency);

-- +goose Down
ALTER TABLE accounts
    DROP CONSTRAINT accounts_merchant_currency_unique,
    DROP CONSTRAINT accounts_merchant_valid,
    DROP COLUMN merchant_id;