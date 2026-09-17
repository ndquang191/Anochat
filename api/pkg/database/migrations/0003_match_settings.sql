ALTER TABLE profiles
ADD COLUMN IF NOT EXISTS match_preference VARCHAR(20);

ALTER TABLE profiles
DROP CONSTRAINT IF EXISTS profiles_match_preference_check;

ALTER TABLE profiles
ADD CONSTRAINT profiles_match_preference_check
CHECK (match_preference IS NULL OR match_preference IN ('mixed', 'opposite_sex'));

CREATE TABLE IF NOT EXISTS match_settings (
    id BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (id),
    default_mode VARCHAR(20) NOT NULL DEFAULT 'mixed'
        CHECK (default_mode IN ('mixed', 'opposite_sex')),
    allow_user_choice BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO match_settings (id, default_mode, allow_user_choice)
VALUES (TRUE, 'mixed', FALSE)
ON CONFLICT (id) DO NOTHING;
