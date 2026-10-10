-- +goose Up
CREATE TABLE operations (
                            id UUID PRIMARY KEY,
                            kind TEXT NOT NULL,
                            request JSONB NOT NULL,
                            request_hash BYTEA NOT NULL,
                            status TEXT NOT NULL DEFAULT 'processing',
                            result JSONB,
                            error_code TEXT,
                            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                            completed_at TIMESTAMPTZ,

                            CONSTRAINT operations_kind_valid
                                CHECK (kind IN ('fund', 'reserve', 'capture', 'release')),

                            CONSTRAINT operations_request_object
                                CHECK (jsonb_typeof(request) = 'object'),

                            CONSTRAINT operations_request_hash_valid
                                CHECK (octet_length(request_hash) = 32),

                            CONSTRAINT operations_result_object
                                CHECK (result IS NULL OR jsonb_typeof(result) = 'object'),

                            CONSTRAINT operations_error_code_valid
                                CHECK (error_code IS NULL OR error_code <> ''),

                            CONSTRAINT operations_state_valid CHECK (
                                (
                                    status = 'processing'
                                        AND result IS NULL
                                        AND error_code IS NULL
                                        AND completed_at IS NULL
                                    )
                                    OR (
                                    status = 'succeeded'
                                        AND result IS NOT NULL
                                        AND error_code IS NULL
                                        AND completed_at IS NOT NULL
                                    )
                                    OR (
                                    status = 'rejected'
                                        AND result IS NULL
                                        AND error_code IS NOT NULL
                                        AND completed_at IS NOT NULL
                                    )
                                )
);

-- +goose Down
DROP TABLE operations;