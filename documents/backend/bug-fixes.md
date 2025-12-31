# Bug Fix: Users Can't Match After Leaving Room

## Problem

After users leave a chat room (or disconnect), they cannot join the matching queue again. The system reports "user already has active room" even though they left the room.

## Root Cause

The WebSocket handler was not updating the database when users leave rooms. Specifically:

1. **When users click "Leave Room"** (`handleLeaveRoom` in `websocket.go:415`):
   - ✅ WebSocket state was updated (removed from room in memory)
   - ✅ Partner was notified via WebSocket
   - ❌ **Database was NEVER updated** - `ended_at` stayed NULL

2. **When users disconnect** (`unregisterClient` in `websocket.go:113`):
   - ✅ WebSocket connections cleaned up
   - ✅ Queue service notified of disconnection
   - ❌ **Database was NEVER updated** - `ended_at` stayed NULL

When users tried to join the queue again, the system checked for active rooms:
```go
// queue.go:136-140
activeRoom, err := qs.userService.GetActiveRoom(ctx, userID)
if err == nil && activeRoom != nil {
    return nil, fmt.Errorf("user already has active room")
}
```

The `GetActiveRoom` query looks for:
```go
// user.go:164
Where("(user1_id = ? OR user2_id = ?) AND ended_at IS NULL AND is_deleted = false", userID, userID)
```

Since `ended_at` was still NULL, the old room was found and users couldn't join the queue.

## Solution

Updated two functions in `api/internal/handler/websocket.go`:

### 1. `handleLeaveRoom` (line 415)
Added database update before WebSocket cleanup:
```go
// Update database - mark room as ended
if err := c.Hub.roomService.LeaveRoom(context.Background(), roomID, c.UserID); err != nil {
    slog.Error("Failed to leave room in database", "error", err, "user_id", c.UserID, "room_id", roomID)
    // Continue with WebSocket cleanup even if DB update fails
}
```

### 2. `unregisterClient` (line 113)
Added database update when users disconnect:
```go
// Update database - mark room as ended
if err := h.roomService.LeaveRoom(context.Background(), *client.RoomID, client.UserID); err != nil {
    slog.Error("Failed to leave room in database on disconnect", "error", err, "user_id", client.UserID, "room_id", *client.RoomID)
    // Continue with cleanup even if DB update fails
}
```

## Testing the Fix

### 1. Restart the backend
```bash
cd api
go run cmd/server/main.go
```

### 2. Test the flow
1. Open two browser windows
2. Have two users match and create a room
3. One user clicks "Leave room" or closes the browser
4. Both users should now be able to join the queue again immediately
5. They should be able to match with new partners

### 3. Expected Behavior
- ✅ Users can leave rooms and rejoin queue immediately
- ✅ Disconnected users are cleaned up properly
- ✅ Room `ended_at` is set when users leave
- ✅ No more "user already has active room" errors

## Cleanup Existing Bad Data

If you have users currently stuck (can't match because of orphaned rooms), run this SQL script:

```bash
# Connect to your PostgreSQL database
psql -U your_username -d anochat_db

# Check for stuck rooms
SELECT id, user1_id, user2_id, created_at,
       EXTRACT(EPOCH FROM (NOW() - created_at))/3600 as hours_old
FROM rooms
WHERE ended_at IS NULL
  AND is_deleted = false
  AND created_at < NOW() - INTERVAL '1 hour';

# Fix stuck rooms (set ended_at to current time)
UPDATE rooms
SET ended_at = NOW()
WHERE ended_at IS NULL
  AND is_deleted = false
  AND created_at < NOW() - INTERVAL '1 hour';
```

Or use the prepared script:
```bash
psql -U your_username -d anochat_db -f api/scripts/cleanup_orphaned_rooms.sql
```

## Files Changed

- `api/internal/handler/websocket.go` - Added `LeaveRoom` database calls
- `api/scripts/cleanup_orphaned_rooms.sql` - SQL script to fix existing bad data (NEW)

## Related Code Locations

- Queue join validation: `api/internal/service/queue.go:136-140`
- Get active room query: `api/internal/service/user.go:159-172`
- Leave room logic: `api/internal/service/room.go:76-99`
