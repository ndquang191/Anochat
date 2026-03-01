package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/matching"
	"github.com/ndquang191/Anochat/api/internal/repository"
)

type QueueService struct {
	roomService *RoomService
	roomRepo    repository.RoomRepository

	mu      sync.Mutex
	entries []*matching.QueueEntry
	inQueue map[uuid.UUID]bool

	matchNotifier matching.MatchNotifier
}

func NewQueueService(roomService *RoomService, roomRepo repository.RoomRepository) *QueueService {
	return &QueueService{
		roomService: roomService,
		roomRepo:    roomRepo,
		inQueue:     make(map[uuid.UUID]bool),
	}
}

func (qs *QueueService) SetMatchNotifier(notifier matching.MatchNotifier) {
	qs.matchNotifier = notifier
}

func (qs *QueueService) JoinQueue(ctx context.Context, userID uuid.UUID) error {
	// Check for active room
	_, err := qs.roomRepo.FindActiveByUserID(ctx, userID)
	if err == nil {
		return fmt.Errorf("bạn đang có phòng chat đang hoạt động, vui lòng rời phòng trước khi tham gia hàng chờ")
	} else if err != repository.ErrNotFound {
		return fmt.Errorf("failed to check active room: %w", err)
	}

	qs.mu.Lock()
	defer qs.mu.Unlock()

	if qs.inQueue[userID] {
		return fmt.Errorf("bạn đã ở trong hàng chờ")
	}

	// If someone is already waiting, match immediately
	for i, entry := range qs.entries {
		if entry.UserID != userID {
			// Remove the waiting entry
			qs.entries = append(qs.entries[:i], qs.entries[i+1:]...)
			delete(qs.inQueue, entry.UserID)

			// Create room
			room, err := qs.roomService.CreateRoom(context.Background(), entry.UserID, userID)
			if err != nil {
				slog.Error("Failed to create room for match", "error", err)
				// Put the entry back
				qs.entries = append(qs.entries, entry)
				qs.inQueue[entry.UserID] = true
				return fmt.Errorf("failed to create room: %w", err)
			}

			slog.Info("Match found", "room_id", room.ID, "user1_id", entry.UserID, "user2_id", userID)

			if qs.matchNotifier != nil {
				qs.matchNotifier.NotifyMatch(entry.UserID, userID, room.ID)
			}
			return nil
		}
	}

	// No match available, add to queue
	entry := &matching.QueueEntry{
		UserID: userID,
	}
	qs.entries = append(qs.entries, entry)
	qs.inQueue[userID] = true

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

func (qs *QueueService) Stop() {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.entries = nil
	qs.inQueue = make(map[uuid.UUID]bool)
}
