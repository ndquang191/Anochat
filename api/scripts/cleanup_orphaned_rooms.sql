-- Cleanup script for orphaned/stuck rooms
-- This script marks rooms as ended if they were created more than 1 hour ago
-- and still have ended_at as NULL (likely due to the bug where LeaveRoom wasn't called)

-- Check how many rooms would be affected
SELECT
    id,
    user1_id,
    user2_id,
    category,
    created_at,
    ended_at,
    is_deleted,
    EXTRACT(EPOCH FROM (NOW() - created_at))/3600 as hours_old
FROM rooms
WHERE ended_at IS NULL
  AND is_deleted = false
  AND created_at < NOW() - INTERVAL '1 hour'
ORDER BY created_at DESC;

-- Uncomment the following to actually update the rooms
-- UPDATE rooms
-- SET ended_at = NOW()
-- WHERE ended_at IS NULL
--   AND is_deleted = false
--   AND created_at < NOW() - INTERVAL '1 hour';

-- Verify the update (run after uncommenting above)
-- SELECT
--     id,
--     user1_id,
--     user2_id,
--     ended_at,
--     EXTRACT(EPOCH FROM (ended_at - created_at))/60 as duration_minutes
-- FROM rooms
-- WHERE ended_at > NOW() - INTERVAL '5 minutes'
-- ORDER BY ended_at DESC;
