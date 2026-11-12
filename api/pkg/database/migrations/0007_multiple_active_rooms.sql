DROP TRIGGER IF EXISTS rooms_active_membership_trigger ON rooms;

ALTER TABLE active_room_members
    DROP CONSTRAINT IF EXISTS active_room_members_pkey;

ALTER TABLE active_room_members
    ADD PRIMARY KEY (user_id, room_id);

CREATE OR REPLACE FUNCTION sync_active_room_members()
RETURNS TRIGGER AS $$
DECLARE
    participant UUID;
    participants UUID[];
BEGIN
    IF TG_OP = 'UPDATE' THEN
        IF OLD.ended_at IS NULL THEN
            DELETE FROM active_room_members WHERE room_id = OLD.id;
        END IF;
    END IF;

    IF NEW.ended_at IS NULL THEN
        SELECT ARRAY_AGG(value ORDER BY value)
        INTO participants
        FROM UNNEST(ARRAY[NEW.user1_id, NEW.user2_id]) AS value;

        FOREACH participant IN ARRAY participants
        LOOP
            -- Serialize room creation per user so concurrent matches cannot
            -- both observe an available third slot.
            PERFORM pg_advisory_xact_lock(hashtextextended(participant::TEXT, 0));
            IF (SELECT COUNT(*) FROM active_room_members WHERE user_id = participant) >= 3 THEN
                RAISE EXCEPTION 'active room limit reached for user %', participant
                    USING ERRCODE = 'check_violation';
            END IF;
            INSERT INTO active_room_members (user_id, room_id, created_at)
            VALUES (participant, NEW.id, NEW.created_at);
        END LOOP;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER rooms_active_membership_trigger
AFTER INSERT OR UPDATE OF ended_at, user1_id, user2_id ON rooms
FOR EACH ROW
EXECUTE FUNCTION sync_active_room_members();
