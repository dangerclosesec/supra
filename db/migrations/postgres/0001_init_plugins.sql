-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS extensions;

GRANT USAGE ON SCHEMA extensions TO public;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA extensions TO public;

ALTER DEFAULT PRIVILEGES IN SCHEMA extensions
    GRANT EXECUTE ON FUNCTIONS TO public;
ALTER DEFAULT PRIVILEGES IN SCHEMA extensions
    GRANT USAGE ON TYPES TO public;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" SCHEMA extensions;
CREATE EXTENSION IF NOT EXISTS citext SCHEMA extensions;
CREATE EXTENSION IF NOT EXISTS pgcrypto SCHEMA extensions;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP EXTENSION IF EXISTS "uuid-ossp" CASCADE;
DROP EXTENSION IF EXISTS citext CASCADE;
DROP EXTENSION IF EXISTS pgcrypto CASCADE;
DROP SCHEMA IF EXISTS extensions CASCADE;
-- +goose StatementEnd
