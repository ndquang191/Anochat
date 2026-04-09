//go:build ignore

package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/internal/domain/matching"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/pkg/metrics"
)

const fakeMatchDelay = 3 * time.Minute

type QueueService struct {
	roomService      *RoomService
	roomRepo         repository.RoomRepository
	profileRepo      repository.ProfileRepository
	fakeMatchService *FakeMatchService

	mu      sync.Mutex
	entries []*matching.QueueEntry
	inQueue map[uuid.UUID]bool

	matchNotifier matching.MatchNotifier
}

func NewQueueService(
	roomService *RoomService,
	roomRepo repository.RoomRepository,
	profileRepo repository.ProfileRepository,
	fakeMatchService *FakeMatchService,
) *QueueService {
	return &QueueService{
		roomService:      roomService,
		roomRepo:         roomRepo,
		profileRepo:      profileRepo,
		fakeMatchService: fakeMatchService,
		inQueue:          make(map[uuid.UUID]bool),
	}
}

func (qs *QueueService) SetMatchNotifier(notifier matching.MatchNotifier) {
	qs.matchNotifier = notifier
}

func (qs *QueueService) JoinQueue(ctx context.Context, userID uuid.UUID) error {
	_, err := qs.roomRepo.FindActiveByUserID(ctx, userID)
	if err == nil {
		return fmt.Errorf("ban dang co phong chat dang hoat dong, vui long roi phong truoc khi tham gia hang cho")
	} else if err != repository.ErrNotFound {
		return fmt.Errorf("failed to check active room: %w", err)
	}

	if qs.fakeMatchService.HasActiveSession(userID) {
		return fmt.Errorf("ban dang co phong chat dang hoat dong, vui long roi phong truoc khi tham gia hang cho")
	}

	qs.mu.Lock()

	if qs.inQueue[userID] {
		qs.mu.Unlock()
		return fmt.Errorf("ban da o trong hang cho")
	}

	for i, entry := range qs.entries {
		if entry.UserID == userID {
			continue
		}

		qs.entries = append(qs.entries[:i], qs.entries[i+1:]...)
		delete(qs.inQueue, entry.UserID)
		metrics.QueueSize.Dec()

		room, err := qs.roomService.CreateRoom(context.Background(), entry.UserID, userID)
		if err != nil {
			slog.Error("Failed to create room for match", "error", err)
			qs.entries = append(qs.entries, entry)
			qs.inQueue[entry.UserID] = true
			metrics.QueueSize.Inc()
			qs.mu.Unlock()
			return fmt.Errorf("failed to create room: %w", err)
		}

		now := time.Now()
		metrics.MatchDuration.Observe(now.Sub(entry.JoinedAt).Seconds())

		slog.Info("Match found", "room_id", room.ID, "user1_id", entry.UserID, "user2_id", userID)

		if qs.matchNotifier != nil {
			qs.matchNotifier.NotifyMatch(entry.UserID, userID, room.ID)
		}
		qs.mu.Unlock()
		return nil
	}

	entry := &matching.QueueEntry{
		UserID:   userID,
		JoinedAt: time.Now(),
	}
	qs.entries = append(qs.entries, entry)
	qs.inQueue[userID] = true
	metrics.QueueSize.Inc()
	qs.mu.Unlock()

	qs.scheduleFakeMatch(entry)

	slog.Info("User joined queue", "user_id", userID, "queue_size", len(qs.entries))
	return nil
}

func (qs *QueueService) LeaveQueue(_ context.Context, userID uuid.UUID) error {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	if !qs.inQueue[userID] {
		return fmt.Errorf("user not found in queue")
	}

	qs.removeUserLocked(userID)
	metrics.QueueSize.Dec()
	slog.Info("User left queue", "user_id", userID, "queue_size", len(qs.entries))
	return nil
}

func (qs *QueueService) IsInQueue(userID uuid.UUID) bool {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	return qs.inQueue[userID]
}

func (qs *QueueService) UserDisconnected(userID uuid.UUID) {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	if qs.inQueue[userID] {
		qs.removeUserLocked(userID)
		metrics.QueueSize.Dec()
		slog.Info("User disconnected, removed from queue", "user_id", userID, "queue_size", len(qs.entries))
	}
}

func (qs *QueueService) removeUserLocked(userID uuid.UUID) {
	for i, entry := range qs.entries {
		if entry.UserID == userID {
			qs.entries = append(qs.entries[:i], qs.entries[i+1:]...)
			break
		}
	}
	delete(qs.inQueue, userID)
}

func (qs *QueueService) scheduleFakeMatch(entry *matching.QueueEntry) {
	go func(userID uuid.UUID, joinedAt time.Time) {
		timer := time.NewTimer(fakeMatchDelay)
		defer timer.Stop()
		<-timer.C

		qs.mu.Lock()
		if !qs.inQueue[userID] {
			qs.mu.Unlock()
			return
		}

		var queuedEntry *matching.QueueEntry
		for _, current := range qs.entries {
			if current.UserID == userID {
				queuedEntry = current
				break
			}
		}

		if queuedEntry == nil || !queuedEntry.JoinedAt.Equal(joinedAt) {
			qs.mu.Unlock()
			return
		}

		qs.removeUserLocked(userID)
		metrics.QueueSize.Dec()
		qs.mu.Unlock()

		profile := qs.loadProfile(userID)
		session := qs.fakeMatchService.CreateSession(userID, profile)

		slog.Info("Created fake match after queue timeout", "user_id", userID, "room_id", session.RoomID)

		if qs.matchNotifier != nil {
			qs.matchNotifier.NotifyFakeMatch(session)
		}

		go qs.finishFakeMatch(userID, session.RoomID)
	}(entry.UserID, entry.JoinedAt)
}

func (qs *QueueService) finishFakeMatch(userID, roomID uuid.UUID) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	profile, err := qs.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil
	}
	return profile
}

func (qs *QueueService) Stop() {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.entries = nil
	qs.inQueue = make(map[uuid.UUID]bool)
}
