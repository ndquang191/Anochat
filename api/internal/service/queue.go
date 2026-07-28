package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/domain/matching"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
	"github.com/ndquang191/Anochat/api/pkg/metrics"
	"github.com/redis/go-redis/v9"
)

const (
	queueKey                   = "queue:waiting"
	queueReservationsKey       = "queue:reservations"
	queueReservationTokensKey  = "queue:reservation_tokens"
	matchReservationTTL        = 30 * time.Second
	matchReservationCleanupTTL = 2 * time.Second
	queueReconcileInterval     = 15 * time.Second
	queueReconcileBatchRooms   = 250
	matchFinalizeAttempts      = 3
)

// matchOrJoinScript atomically reserves a waiting partner for the joining user,
// or adds the joining user to the queue if no partner is available. A reserved
// partner stays in the waiting queue until room creation succeeds, but other
// match attempts skip them.
//
// Returns:
//
//	"match_in_progress"              - user already owns or belongs to a live reservation
//	"already_in_queue"               - user is already waiting
//	"reserved:<partnerID>:<score>"   - partner reserved until the lease expires
//	"waiting"                        - no match; user added to queue
var matchOrJoinScript = redis.NewScript(`
local queue = KEYS[1]
local reservations = KEYS[2]
local reservationTokens = KEYS[3]
local userID = ARGV[1]
local score = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local expiresAt = tonumber(ARGV[4])
local token = ARGV[5]

local expiredUsers = redis.call("ZRANGEBYSCORE", reservations, "-inf", now)
for _, expiredUser in ipairs(expiredUsers) do
    redis.call("ZREM", reservations, expiredUser)
    redis.call("HDEL", reservationTokens, expiredUser)
end

if redis.call("ZSCORE", reservations, userID) ~= false then
    return "match_in_progress"
end

if redis.call("ZSCORE", queue, userID) ~= false then
    return "already_in_queue"
end

local candidates = redis.call("ZRANGE", queue, 0, -1, "WITHSCORES")
for index = 1, #candidates, 2 do
    local candidate = candidates[index]
    local partnerScore = candidates[index + 1]
    if candidate ~= userID and redis.call("ZSCORE", reservations, candidate) == false then
        redis.call("ZADD", reservations, expiresAt, userID, expiresAt, candidate)
        redis.call("HSET", reservationTokens, userID, token, candidate, token)
        return "reserved:" .. candidate .. ":" .. partnerScore
    end
end

redis.call("ZADD", queue, score, userID)
return "waiting"
`)

// finishMatchReservationScript only commits or releases a reservation when the
// supplied token still owns both users. This prevents a late cleanup from
// deleting a newer reservation created after the original lease expired.
var finishMatchReservationScript = redis.NewScript(`
local queue = KEYS[1]
local reservations = KEYS[2]
local reservationTokens = KEYS[3]
local userID = ARGV[1]
local partnerID = ARGV[2]
local token = ARGV[3]
local action = ARGV[4]

if redis.call("HGET", reservationTokens, userID) ~= token
    or redis.call("HGET", reservationTokens, partnerID) ~= token then
    return 0
end

if action == "commit" then
    redis.call("ZREM", queue, userID, partnerID)
end

redis.call("ZREM", reservations, userID, partnerID)
redis.call("HDEL", reservationTokens, userID, partnerID)
return 1
`)

// removeFromQueueScript keeps a user in place while a live match reservation
// owns them. This closes the race where leave/disconnect removed a partner
// after matching started but before the room transaction committed.
var removeFromQueueScript = redis.NewScript(`
local queue = KEYS[1]
local reservations = KEYS[2]
local reservationTokens = KEYS[3]
local userID = ARGV[1]
local now = tonumber(ARGV[2])

local expiredUsers = redis.call("ZRANGEBYSCORE", reservations, "-inf", now)
for _, expiredUser in ipairs(expiredUsers) do
    redis.call("ZREM", reservations, expiredUser)
    redis.call("HDEL", reservationTokens, expiredUser)
end

if redis.call("ZSCORE", reservations, userID) ~= false then
    return -1
end

return redis.call("ZREM", queue, userID)
`)

// reconcileActiveRoomsScript treats PostgreSQL's active-room membership as the
// source of truth and removes stale queue/reservation state left by a failed
// Redis finalization.
var reconcileActiveRoomsScript = redis.NewScript(`
local queue = KEYS[1]
local reservations = KEYS[2]
local reservationTokens = KEYS[3]

local removed = 0
for index = 1, #ARGV, 3 do
    local userID = ARGV[index]
    local roomCreatedAt = tonumber(ARGV[index + 1])
    local reservationCutoff = tonumber(ARGV[index + 2])

    local queueScore = redis.call("ZSCORE", queue, userID)
    if queueScore ~= false and tonumber(queueScore) <= roomCreatedAt then
        removed = removed + redis.call("ZREM", queue, userID)
    end

    local reservationExpiry = redis.call("ZSCORE", reservations, userID)
    if reservationExpiry ~= false and tonumber(reservationExpiry) < reservationCutoff then
        redis.call("ZREM", reservations, userID)
        redis.call("HDEL", reservationTokens, userID)
    end
end
return removed
`)

type QueueService struct {
	roomService   *RoomService
	roomRepo      repository.RoomRepository
	rdb           *redis.Client
	matchNotifier matching.MatchNotifier
}

func NewQueueService(
	roomService *RoomService,
	roomRepo repository.RoomRepository,
	rdb *redis.Client,
) *QueueService {
	return &QueueService{
		roomService: roomService,
		roomRepo:    roomRepo,
		rdb:         rdb,
	}
}

func (qs *QueueService) SetMatchNotifier(notifier matching.MatchNotifier) {
	qs.matchNotifier = notifier
}

func (qs *QueueService) JoinQueue(ctx context.Context, userID uuid.UUID) error {
	activeRoom, err := qs.roomRepo.FindActiveByUserID(ctx, userID)
	if err == nil {
		if reconcileErr := qs.reconcileRooms(ctx, []*chat.Room{activeRoom}); reconcileErr != nil {
			slog.Warn("Failed to reconcile active room before rejecting queue join", "room_id", activeRoom.ID, "error", reconcileErr)
		}
		return apperr.ErrHasActiveRoom
	} else if err != repository.ErrNotFound {
		return fmt.Errorf("failed to check active room: %w", err)
	}

	now := time.Now()
	score := float64(now.UnixMilli())
	reservationToken := uuid.NewString()
	result, err := matchOrJoinScript.Run(
		ctx,
		qs.rdb,
		[]string{queueKey, queueReservationsKey, queueReservationTokensKey},
		userID.String(),
		score,
		now.UnixMilli(),
		now.Add(matchReservationTTL).UnixMilli(),
		reservationToken,
	).Text()
	if err != nil {
		return fmt.Errorf("queue script error: %w", err)
	}

	switch {
	case result == "match_in_progress":
		return apperr.ErrMatchInProgress

	case result == "already_in_queue":
		return apperr.ErrAlreadyInQueue

	case strings.HasPrefix(result, "reserved:"):
		parts := strings.SplitN(strings.TrimPrefix(result, "reserved:"), ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("unexpected queue script result: %s", result)
		}

		partnerID, err := uuid.Parse(parts[0])
		if err != nil {
			return fmt.Errorf("invalid partner ID from queue: %w", err)
		}

		partnerJoinMs, _ := strconv.ParseFloat(parts[1], 64)
		room, err := qs.roomService.CreateRoom(ctx, partnerID, userID)
		if err != nil {
			if released, releaseErr := qs.finishMatchReservationWithRetry(userID, partnerID, reservationToken, "release"); releaseErr != nil {
				slog.Error("Failed to release match reservation", "error", releaseErr, "user1", partnerID, "user2", userID)
			} else if !released {
				slog.Warn("Match reservation expired before release", "user1", partnerID, "user2", userID)
			}
			slog.Error("Failed to create room for match", "error", err, "user1", partnerID, "user2", userID)
			return fmt.Errorf("failed to create room: %w", err)
		}

		if committed, commitErr := qs.finishMatchReservationWithRetry(userID, partnerID, reservationToken, "commit"); commitErr != nil {
			slog.Error("Failed to commit match reservation", "error", commitErr, "room_id", room.ID, "user1", partnerID, "user2", userID)
		} else if !committed {
			slog.Warn("Match reservation expired before commit", "room_id", room.ID, "user1", partnerID, "user2", userID)
			if reconcileErr := qs.reconcileRooms(context.Background(), []*chat.Room{room}); reconcileErr != nil {
				slog.Error("Failed to reconcile expired match reservation", "error", reconcileErr, "room_id", room.ID)
			}
		}

		waitSeconds := time.Since(time.UnixMilli(int64(partnerJoinMs))).Seconds()
		metrics.MatchDuration.Observe(waitSeconds)

		qs.updateQueueMetric(ctx)
		slog.Info("Match found", "room_id", room.ID, "user1_id", partnerID, "user2_id", userID, "wait_seconds", waitSeconds)

		if qs.matchNotifier != nil {
			qs.matchNotifier.NotifyMatch(partnerID, userID, room.ID)
		}

	case result == "waiting":
		qs.updateQueueMetric(ctx)
		slog.Info("User joined queue", "user_id", userID)
	}

	return nil
}

func (qs *QueueService) finishMatchReservationWithRetry(
	userID uuid.UUID,
	partnerID uuid.UUID,
	token string,
	action string,
) (bool, error) {
	var lastErr error
	for attempt := 0; attempt < matchFinalizeAttempts; attempt++ {
		finished, err := qs.finishMatchReservation(userID, partnerID, token, action)
		if err == nil {
			return finished, nil
		}
		lastErr = err
		if attempt+1 < matchFinalizeAttempts {
			time.Sleep(time.Duration(attempt+1) * 50 * time.Millisecond)
		}
	}
	return false, lastErr
}

func (qs *QueueService) finishMatchReservation(
	userID uuid.UUID,
	partnerID uuid.UUID,
	token string,
	action string,
) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), matchReservationCleanupTTL)
	defer cancel()

	result, err := finishMatchReservationScript.Run(
		ctx,
		qs.rdb,
		[]string{queueKey, queueReservationsKey, queueReservationTokensKey},
		userID.String(),
		partnerID.String(),
		token,
		action,
	).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func (qs *QueueService) LeaveQueue(ctx context.Context, userID uuid.UUID) error {
	removed, err := removeFromQueueScript.Run(
		ctx,
		qs.rdb,
		[]string{queueKey, queueReservationsKey, queueReservationTokensKey},
		userID.String(),
		time.Now().UnixMilli(),
	).Int()
	if err != nil {
		return fmt.Errorf("failed to leave queue: %w", err)
	}
	if removed == -1 {
		return apperr.ErrMatchInProgress
	}
	if removed == 0 {
		return apperr.ErrNotInQueue
	}

	qs.updateQueueMetric(ctx)
	slog.Info("User left queue", "user_id", userID)
	return nil
}

func (qs *QueueService) IsInQueue(userID uuid.UUID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	score, err := qs.rdb.ZScore(ctx, queueKey, userID.String()).Result()
	return err == nil && score >= 0
}

func (qs *QueueService) UserDisconnected(userID uuid.UUID) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	removed, err := removeFromQueueScript.Run(
		ctx,
		qs.rdb,
		[]string{queueKey, queueReservationsKey, queueReservationTokensKey},
		userID.String(),
		time.Now().UnixMilli(),
	).Int()
	if err != nil {
		slog.Warn("Failed to remove disconnected user from queue", "user_id", userID, "error", err)
		return
	}
	if removed == -1 {
		slog.Info("User disconnected while match is in progress; reservation retained", "user_id", userID)
		return
	}
	if removed > 0 {
		qs.updateQueueMetric(ctx)
		slog.Info("User disconnected, removed from queue", "user_id", userID)
	}
}

func (qs *QueueService) Run(ctx context.Context) {
	qs.reconcileAllActiveRooms(ctx)
	ticker := time.NewTicker(queueReconcileInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			qs.reconcileAllActiveRooms(ctx)
		}
	}
}

func (qs *QueueService) reconcileAllActiveRooms(ctx context.Context) {
	rooms, err := qs.roomRepo.ListActive(ctx)
	if err != nil {
		slog.Error("Failed to list active rooms for queue reconciliation", "error", err)
		return
	}
	if err := qs.reconcileRooms(ctx, rooms); err != nil {
		slog.Error("Failed to reconcile active rooms with Redis queue", "error", err)
	}
}

func (qs *QueueService) reconcileRooms(ctx context.Context, rooms []*chat.Room) error {
	if len(rooms) == 0 {
		return nil
	}
	for start := 0; start < len(rooms); start += queueReconcileBatchRooms {
		end := min(start+queueReconcileBatchRooms, len(rooms))
		entries := make([]interface{}, 0, (end-start)*6)
		for _, room := range rooms[start:end] {
			createdAt := room.CreatedAt.UnixMilli()
			reservationCutoff := room.CreatedAt.Add(matchReservationTTL).UnixMilli()
			entries = append(
				entries,
				room.User1ID.String(), createdAt, reservationCutoff,
				room.User2ID.String(), createdAt, reservationCutoff,
			)
		}
		if _, err := reconcileActiveRoomsScript.Run(
			ctx,
			qs.rdb,
			[]string{queueKey, queueReservationsKey, queueReservationTokensKey},
			entries...,
		).Result(); err != nil {
			return err
		}
	}
	qs.updateQueueMetric(ctx)
	return nil
}

func (qs *QueueService) updateQueueMetric(ctx context.Context) {
	size, err := qs.rdb.ZCard(ctx, queueKey).Result()
	if err == nil {
		metrics.QueueSize.Set(float64(size))
	}
}
