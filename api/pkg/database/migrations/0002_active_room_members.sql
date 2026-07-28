DO $$
BEGIN
    IF EXISTS (
        SELECT user_id
        FROM (
            SELECT user1_id AS user_id FROM rooms WHERE ended_at IS NULL
            UNION ALL
            SELECT user2_id AS user_id FROM rooms WHERE ended_at IS NULL
        ) active_users
        GROUP BY user_id
        HAVING COUNT(*) > 1
    ) THEN
        RAISE EXCEPTION 'cannot enforce one active room per user: duplicate active-room memberships exist';
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS active_room_members (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_active_room_members_room_id
    ON active_room_members(room_id);

INSERT INTO active_room_members (user_id, room_id, created_at)
SELECT user1_id, id, created_at
FROM rooms
WHERE ended_at IS NULL
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO active_room_members (user_id, room_id, created_at)
SELECT user2_id, id, created_at
FROM rooms
WHERE ended_at IS NULL
ON CONFLICT (user_id) DO NOTHING;

CREATE OR REPLACE FUNCTION sync_active_room_members()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.ended_at IS NULL THEN
            INSERT INTO active_room_members (user_id, room_id, created_at)
            VALUES
                (NEW.user1_id, NEW.id, NEW.created_at),
                (NEW.user2_id, NEW.id, NEW.created_at);
        END IF;
        RETURN NEW;
    END IF;

    IF OLD.ended_at IS NULL THEN
        DELETE FROM active_room_members WHERE room_id = OLD.id;
    END IF;

    IF NEW.ended_at IS NULL THEN
        INSERT INTO active_room_members (user_id, room_id, created_at)
        VALUES
            (NEW.user1_id, NEW.id, NEW.created_at),
            (NEW.user2_id, NEW.id, NEW.created_at);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS rooms_active_membership_trigger ON rooms;
CREATE TRIGGER rooms_active_membership_trigger
AFTER INSERT OR UPDATE OF ended_at, user1_id, user2_id ON rooms
FOR EACH ROW
EXECUTE FUNCTION sync_active_room_members();
