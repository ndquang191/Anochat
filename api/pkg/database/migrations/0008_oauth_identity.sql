ALTER TABLE users
    ADD COLUMN IF NOT EXISTS oauth_provider TEXT,
    ADD COLUMN IF NOT EXISTS oauth_subject TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_oauth_identity
    ON users(oauth_provider, oauth_subject)
    WHERE oauth_provider IS NOT NULL AND oauth_subject IS NOT NULL;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_oauth_identity_complete;

ALTER TABLE users
    ADD CONSTRAINT users_oauth_identity_complete CHECK (
        (oauth_provider IS NULL AND oauth_subject IS NULL)
        OR (oauth_provider IS NOT NULL AND oauth_subject IS NOT NULL)
    );
