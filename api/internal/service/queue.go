package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/internal/domain/matching"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
	"github.com/ndquang191/Anochat/api/pkg/metrics"
	"github.com/redis/go-redis/v9"
)

const (
	queueKey       = "queue:waiting"
	fakeMatchDelay = 3 * time.Minute
)

// matchOrJoinScript atomically tries to match the joining user with a waiting one,
// or adds them to the queue if no match is available.
//
// Returns:
//
//	"already_in_queue"       - user is already waiting
//	"matched:<partnerID>"    - user was matched with partnerID (partner removed from queue)
//	"waiting"                - no match; user added to queue
var matchOrJoinScript = redis.NewScript(`
local queue = KEYS[1]
local userID = ARGV[1]
local score  = tonumber(ARGV[2])

if redis.call("ZSCORE", queue, userID) ~= false then
    return "already_in_queue"
end

local candidates = redis.call("ZRANGE", queue, 0, -1)
for _, candidate in ipairs(candidates) do
    if candidate ~= userID then
        local partnerScore = redis.call("ZSCORE", queue, candidate)
        redis.call("ZREM", queue, candidate)
        return "matched:" .. candidate .. ":" .. partnerScore
    end
end

redis.call("ZADD", queue, score, userID)
return "waiting"
`)

type QueueService struct {
	roomService      *RoomService
	roomRepo         repository.RoomRepository
	profileRepo      repository.ProfileRepository
	fakeMatchService *FakeMatchService
	rdb              *redis.Client
	matchNotifier    matching.MatchNotifier
}

func NewQueueService(
	roomService *RoomService,
	roomRepo repository.RoomRepository,
	rdb *redis.Client,
	extras ...interface{},
) *QueueService {
	qs := &QueueService{
		roomService: roomService,
		roomRepo:    roomRepo,
		rdb:         rdb,
	}

	if len(extras) > 0 {
		if profileRepo, ok := extras[0].(repository.ProfileRepository); ok {
			qs.profileRepo = profileRepo
		}
	}
	if len(extras) > 1 {
		if fakeMatchService, ok := extras[1].(*FakeMatchService); ok {
			qs.fakeMatchService = fakeMatchService
		}
	}

	return qs
}

func (qs *QueueService) SetMatchNotifier(notifier matching.MatchNotifier) {
	qs.matchNotifier = notifier
}

func (qs *QueueService) JoinQueue(ctx context.Context, userID uuid.UUID) error {
	_, err := qs.roomRepo.FindActiveByUserID(ctx, userID)
	if err == nil {
		return apperr.ErrHasActiveRoom
	} else if err != repository.ErrNotFound {
		return fmt.Errorf("failed to check active room: %w", err)
	}

	if qs.fakeMatchService != nil && qs.fakeMatchService.HasActiveSession(userID) {
		return apperr.ErrHasActiveRoom
	}

	score := float64(time.Now().UnixMilli())
	result, err := matchOrJoinScript.Run(ctx, qs.rdb, []string{queueKey}, userID.String(), score).Text()
	if err != nil {
		return fmt.Errorf("queue script error: %w", err)
	}

	switch {
	case result == "already_in_queue":
		return apperr.ErrAlreadyInQueue

	case strings.HasPrefix(result, "matched:"):
		parts := strings.SplitN(strings.TrimPrefix(result, "matched:"), ":", 2)
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
			qs.rdb.ZAdd(ctx, queueKey, redis.Z{Score: partnerJoinMs, Member: partnerID.String()})
			slog.Error("Failed to create room for match", "error", err, "user1", partnerID, "user2", userID)
			return fmt.Errorf("failed to create room: %w", err)
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
		qs.scheduleFakeMatch(userID, int64(score))
	}

	return nil
}

func (qs *QueueService) LeaveQueue(ctx context.Context, userID uuid.UUID) error {
	removed, err := qs.rdb.ZRem(ctx, queueKey, userID.String()).Result()
	if err != nil {
		return fmt.Errorf("failed to leave queue: %w", err)
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

	removed, err := qs.rdb.ZRem(ctx, queueKey, userID.String()).Result()
	if err != nil {
		slog.Warn("Failed to remove disconnected user from queue", "user_id", userID, "error", err)
		return
	}
	if removed > 0 {
		qs.updateQueueMetric(ctx)
		slog.Info("User disconnected, removed from queue", "user_id", userID)
	}
}

// Stop is a no-op for the distributed Redis queue.
// Users are removed individually as they disconnect; clearing the shared
// queue key on one instance shutdown would affect other running instances.
func (qs *QueueService) Stop() {}

func (qs *QueueService) scheduleFakeMatch(userID uuid.UUID, joinedAtMs int64) {
	if qs.fakeMatchService == nil {
		return
	}

	go func() {
		timer := time.NewTimer(fakeMatchDelay)
		defer timer.Stop()
		<-timer.C

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		score, err := qs.rdb.ZScore(ctx, queueKey, userID.String()).Result()
		if err != nil || int64(score) != joinedAtMs {
			return
		}

		removed, err := qs.rdb.ZRem(ctx, queueKey, userID.String()).Result()
		if err != nil || removed == 0 {
			return
		}
		qs.updateQueueMetric(ctx)

		profile := qs.loadProfile(userID)
		session := qs.fakeMatchService.CreateSession(userID, profile)

		slog.Info("Created fake match after queue timeout", "user_id", userID, "room_id", session.RoomID)

		if qs.matchNotifier != nil {
			qs.matchNotifier.NotifyFakeMatch(session)
		}

		go qs.finishFakeMatch(userID, session.RoomID)
	}()
}

func (qs *QueueService) finishFakeMatch(userID, roomID uuid.UUID) {
	if qs.fakeMatchService == nil {
		return
	}

	<-time.After(2 * time.Second)

	session := qs.fakeMatchService.GetByUserID(userID)
	if session == nil || session.RoomID != roomID {
		return
	}

	qs.fakeMatchService.EndSession(userID)
	if qs.matchNotifier != nil {
		qs.matchNotifier.NotifyFakePartnerLeft(userID, roomID)
	}
}

func (qs *QueueService) loadProfile(userID uuid.UUID) *identity.Profile {
	if qs.profileRepo == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	profile, err := qs.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil
	}
	return profile
}

func (qs *QueueService) updateQueueMetric(ctx context.Context) {
	size, err := qs.rdb.ZCard(ctx, queueKey).Result()
	if err == nil {
		metrics.QueueSize.Set(float64(size))
	}
}
