ALTER TABLE profiles
ADD COLUMN IF NOT EXISTS birth_year INTEGER;

UPDATE profiles
SET birth_year = EXTRACT(YEAR FROM CURRENT_DATE)::INTEGER - age
WHERE birth_year IS NULL AND age IS NOT NULL;

DROP INDEX IF EXISTS idx_profiles_age;
CREATE INDEX IF NOT EXISTS idx_profiles_birth_year ON profiles(birth_year);

ALTER TABLE profiles
DROP COLUMN IF EXISTS age;
