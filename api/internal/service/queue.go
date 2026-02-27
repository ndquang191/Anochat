package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/matching"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/pkg/config"
)

type genderQueue struct {
	entries []*matching.QueueEntry
}

type QueueService struct {
	roomService *RoomService
	userService *UserService
	roomRepo    repository.RoomRepository
	config      *config.Config

	queues     map[string]map[matching.Gender]*genderQueue
	queueMutex sync.RWMutex

	userConnections map[uuid.UUID]bool
	connMutex       sync.RWMutex

	userPositions map[uuid.UUID]*matching.QueuePosition
	posMutex      sync.RWMutex

	matchStats struct {
		sync.RWMutex
		totalMatches   int64
		maleWaitTime   time.Duration
		femaleWaitTime time.Duration
		lastMatchTime  time.Time
	}

	matchNotifier matching.MatchNotifier
}

func NewQueueService(roomService *RoomService, userService *UserService, roomRepo repository.RoomRepository, cfg *config.Config) *QueueService {
	qs := &QueueService{
		roomService:     roomService,
		userService:     userService,
		roomRepo:        roomRepo,
		config:          cfg,
		queues:          make(map[string]map[matching.Gender]*genderQueue),
		userConnections: make(map[uuid.UUID]bool),
		userPositions:   make(map[uuid.UUID]*matching.QueuePosition),
	}

	slog.Info("QueueService initialized successfully")
	go qs.startCleanupRoutine()
	return qs
}

func (qs *QueueService) SetMatchNotifier(notifier matching.MatchNotifier) {
	qs.matchNotifier = notifier
}

func (qs *QueueService) getOrCreateQueue(category string, gender matching.Gender) *genderQueue {
	if qs.queues[category] == nil {
		qs.queues[category] = make(map[matching.Gender]*genderQueue)
	}
	if qs.queues[category][gender] == nil {
		qs.queues[category][gender] = &genderQueue{}
	}
	return qs.queues[category][gender]
}

func (qs *QueueService) JoinQueue(ctx context.Context, userID uuid.UUID, category string) (*matching.QueueEntry, error) {
	if !config.IsValidCategory(category) {
		return nil, fmt.Errorf("invalid category: %s", category)
	}

	profile, err := qs.userService.GetProfile(ctx, userID)
	if err != nil {
		slog.Error("Failed to get user profile", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	if profile == nil {
		return nil, fmt.Errorf("vui lòng hoàn thành thông tin cá nhân trước khi tham gia hàng chờ")
	}

	_, err = qs.roomRepo.FindActiveByUserID(ctx, userID)
	if err == nil {
		return nil, fmt.Errorf("bạn đang có phòng chat đang hoạt động, vui lòng rời phòng trước khi tham gia hàng chờ")
	} else if err != repository.ErrNotFound {
		return nil, fmt.Errorf("failed to check active room: %w", err)
	}

	if qs.isUserInQueue(userID) {
		return nil, fmt.Errorf("bạn đã ở trong hàng chờ")
	}

	gender := matching.GenderFromProfile(profile)

	now := time.Now()
	entry := &matching.QueueEntry{
		UserID:    userID,
		Profile:   profile,
		Gender:    gender,
		Category:  category,
		JoinedAt:  now,
		ExpiresAt: now.Add(qs.config.Chat.QueueHeartbeatTTL),
		IsMatched: false,
		MatchChan: make(chan *matching.MatchResult, 1),
	}

	qs.queueMutex.Lock()
	q := qs.getOrCreateQueue(category, gender)
	position := len(q.entries) + 1
	q.entries = append(q.entries, entry)
	qs.queueMutex.Unlock()

	qs.posMutex.Lock()
	qs.userPositions[userID] = &matching.QueuePosition{
		UserID:    userID,
		Position:  position,
		Category:  category,
		Gender:    gender,
		JoinedAt:  now,
		ExpiresAt: entry.ExpiresAt,
	}
	qs.posMutex.Unlock()

	qs.connMutex.Lock()
	qs.userConnections[userID] = true
	qs.connMutex.Unlock()

	slog.Info("User joined queue",
		"user_id", userID,
		"category", category,
		"gender", gender.String(),
		"position", position)

	go qs.tryMatch(entry)
	qs.logCurrentQueueStatus(category)

	return entry, nil
}

func (qs *QueueService) tryMatch(entry *matching.QueueEntry) {
	qs.queueMutex.Lock()
	defer qs.queueMutex.Unlock()

	if entry.IsMatched || !qs.isUserConnected(entry.UserID) {
		return
	}

	category := entry.Category
	match := qs.findMatch(category, entry)

	slog.Info("Match attempt",
		"user_id", entry.UserID,
		"category", category,
		"gender", entry.Gender.String(),
		"found_match", match != nil)

	if match == nil {
		return
	}

	entry.IsMatched = true
	match.IsMatched = true

	qs.removeEntryFromQueue(category, entry)
	qs.removeEntryFromQueue(category, match)
	qs.updateMatchStats(entry, match)

	room, err := qs.roomService.CreateRoom(context.Background(), entry.UserID, match.UserID, category)
	if err != nil {
		entry.MatchChan <- &matching.MatchResult{Error: err}
		match.MatchChan <- &matching.MatchResult{Error: err}
		return
	}

	result := &matching.MatchResult{
		RoomID:   room.ID,
		User1ID:  entry.UserID,
		User2ID:  match.UserID,
		Category: category,
	}
	entry.MatchChan <- result
	match.MatchChan <- result

	slog.Info("Match found",
		"room_id", room.ID,
		"user1_id", entry.UserID,
		"user2_id", match.UserID,
		"category", category)

	if qs.matchNotifier != nil {
		qs.matchNotifier.NotifyMatch(entry.UserID, match.UserID, room.ID, category)
	}

	qs.logCurrentQueueStatus(category)
}

func (qs *QueueService) findMatch(category string, entry *matching.QueueEntry) *matching.QueueEntry {
	catQueues := qs.queues[category]
	if catQueues == nil {
		return nil
	}

	var searchOrder []matching.Gender
	switch entry.Gender {
	case matching.GenderMale:
		searchOrder = []matching.Gender{matching.GenderFemale, matching.GenderMale}
	case matching.GenderFemale:
		searchOrder = []matching.Gender{matching.GenderMale, matching.GenderFemale}
	default:
		searchOrder = []matching.Gender{matching.GenderMale, matching.GenderFemale, matching.GenderUnknown}
	}

	for _, g := range searchOrder {
		if q := catQueues[g]; q != nil {
			for _, other := range q.entries {
				if !other.IsMatched && other.UserID != entry.UserID && qs.isUserConnected(other.UserID) {
					return other
				}
			}
		}
	}
	return nil
}

func (qs *QueueService) removeEntryFromQueue(category string, entry *matching.QueueEntry) {
	catQueues := qs.queues[category]
	if catQueues == nil {
		return
	}
	q := catQueues[entry.Gender]
	if q == nil {
		return
	}

	for i, e := range q.entries {
		if e.UserID == entry.UserID {
			q.entries = append(q.entries[:i], q.entries[i+1:]...)

			qs.posMutex.Lock()
			delete(qs.userPositions, entry.UserID)
			qs.posMutex.Unlock()

			qs.connMutex.Lock()
			delete(qs.userConnections, entry.UserID)
			qs.connMutex.Unlock()

			qs.updatePositionsForQueue(category, entry.Gender, i)
			break
		}
	}
}

func (qs *QueueService) updatePositionsForQueue(category string, gender matching.Gender, fromIndex int) {
	catQueues := qs.queues[category]
	if catQueues == nil {
		return
	}
	q := catQueues[gender]
	if q == nil {
		return
	}

	qs.posMutex.Lock()
	defer qs.posMutex.Unlock()

	for i := fromIndex; i < len(q.entries); i++ {
		if pos, exists := qs.userPositions[q.entries[i].UserID]; exists {
			pos.Position = i + 1
		}
	}
}

func (qs *QueueService) GetQueueStatus(ctx context.Context, userID uuid.UUID) (*matching.QueueStatus, error) {
	qs.posMutex.RLock()
	position, exists := qs.userPositions[userID]
	qs.posMutex.RUnlock()

	if !exists {
		return &matching.QueueStatus{IsInQueue: false}, nil
	}

	return &matching.QueueStatus{
		IsInQueue: true,
		Position:  position.Position,
		Category:  position.Category,
		JoinedAt:  position.JoinedAt,
		ExpiresAt: position.ExpiresAt,
	}, nil
}

func (qs *QueueService) GetMatchStatistics() map[string]interface{} {
	qs.matchStats.RLock()
	defer qs.matchStats.RUnlock()

	return map[string]interface{}{
		"total_matches":    qs.matchStats.totalMatches,
		"male_wait_time":   qs.matchStats.maleWaitTime,
		"female_wait_time": qs.matchStats.femaleWaitTime,
		"last_match_time":  qs.matchStats.lastMatchTime,
	}
}

func (qs *QueueService) GetQueueStats() map[string]map[string]interface{} {
	qs.queueMutex.RLock()
	defer qs.queueMutex.RUnlock()

	stats := make(map[string]map[string]interface{})
	for category, genderQueues := range qs.queues {
		maleCount, femaleCount := 0, 0
		if q := genderQueues[matching.GenderMale]; q != nil {
			maleCount = len(q.entries)
		}
		if q := genderQueues[matching.GenderFemale]; q != nil {
			femaleCount = len(q.entries)
		}
		stats[category] = map[string]interface{}{
			"male_count":          maleCount,
			"female_count":        femaleCount,
			"total_count":         maleCount + femaleCount,
			"estimated_wait_time": qs.calculateEstimatedWaitTime(maleCount, femaleCount),
		}
	}
	return stats
}

func (qs *QueueService) calculateEstimatedWaitTime(maleCount, femaleCount int) time.Duration {
	if maleCount == 0 || femaleCount == 0 {
		return 5 * time.Minute
	}
	diff := maleCount - femaleCount
	if diff < 0 {
		diff = -diff
	}
	if diff <= 2 {
		return 90 * time.Second
	}
	larger := maleCount
	if femaleCount > larger {
		larger = femaleCount
	}
	return time.Duration(larger) * 30 * time.Second
}

func (qs *QueueService) LeaveQueue(ctx context.Context, userID uuid.UUID) error {
	qs.queueMutex.Lock()
	defer qs.queueMutex.Unlock()

	for category, genderQueues := range qs.queues {
		for gender, q := range genderQueues {
			for i, entry := range q.entries {
				if entry.UserID == userID {
					q.entries = append(q.entries[:i], q.entries[i+1:]...)
					close(entry.MatchChan)

					qs.updatePositionsForQueue(category, gender, i)

					qs.connMutex.Lock()
					qs.userConnections[userID] = false
					qs.connMutex.Unlock()

					qs.posMutex.Lock()
					delete(qs.userPositions, userID)
					qs.posMutex.Unlock()

					slog.Info("User manually left queue", "user_id", userID, "category", category, "gender", gender.String())
					qs.logCurrentQueueStatus(category)
					return nil
				}
			}
		}
	}
	return fmt.Errorf("user not found in queue")
}

func (qs *QueueService) UserDisconnected(userID uuid.UUID) {
	qs.connMutex.Lock()
	qs.userConnections[userID] = false
	qs.connMutex.Unlock()

	qs.queueMutex.Lock()
	defer qs.queueMutex.Unlock()

	for category, genderQueues := range qs.queues {
		for gender, q := range genderQueues {
			for i, entry := range q.entries {
				if entry.UserID == userID {
					q.entries = append(q.entries[:i], q.entries[i+1:]...)
					close(entry.MatchChan)
					qs.updatePositionsForQueue(category, gender, i)

					qs.posMutex.Lock()
					delete(qs.userPositions, userID)
					qs.posMutex.Unlock()

					slog.Info("User disconnected from queue", "user_id", userID, "category", category, "gender", gender.String())
					qs.logCurrentQueueStatus(category)
					return
				}
			}
		}
	}
}

func (qs *QueueService) isUserConnected(userID uuid.UUID) bool {
	qs.connMutex.RLock()
	defer qs.connMutex.RUnlock()
	return qs.userConnections[userID]
}

func (qs *QueueService) isUserInQueue(userID uuid.UUID) bool {
	qs.posMutex.RLock()
	defer qs.posMutex.RUnlock()
	_, exists := qs.userPositions[userID]
	return exists
}

func (qs *QueueService) updateMatchStats(entry1, entry2 *matching.QueueEntry) {
	qs.matchStats.Lock()
	defer qs.matchStats.Unlock()

	qs.matchStats.totalMatches++
	qs.matchStats.lastMatchTime = time.Now()

	waitTime1 := time.Since(entry1.JoinedAt)
	waitTime2 := time.Since(entry2.JoinedAt)

	if entry1.Gender == matching.GenderMale {
		qs.matchStats.maleWaitTime = (qs.matchStats.maleWaitTime + waitTime1) / 2
		qs.matchStats.femaleWaitTime = (qs.matchStats.femaleWaitTime + waitTime2) / 2
	} else {
		qs.matchStats.maleWaitTime = (qs.matchStats.maleWaitTime + waitTime2) / 2
		qs.matchStats.femaleWaitTime = (qs.matchStats.femaleWaitTime + waitTime1) / 2
	}
}

func (qs *QueueService) logCurrentQueueStatus(category string) {
	qs.queueMutex.RLock()
	defer qs.queueMutex.RUnlock()

	catQueues := qs.queues[category]
	if catQueues == nil {
		slog.Info("Queue empty", "category", category)
		return
	}

	maleCount, femaleCount := 0, 0
	if q := catQueues[matching.GenderMale]; q != nil {
		maleCount = len(q.entries)
	}
	if q := catQueues[matching.GenderFemale]; q != nil {
		femaleCount = len(q.entries)
	}

	slog.Info("Queue summary",
		"category", category,
		"male_count", maleCount,
		"female_count", femaleCount,
		"total_count", maleCount+femaleCount)
}

func (qs *QueueService) startCleanupRoutine() {
	ticker := time.NewTicker(qs.config.Chat.QueueCleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		qs.cleanupExpiredEntries()
	}
}

func (qs *QueueService) cleanupExpiredEntries() {
	now := time.Now()
	var expiredUsers []uuid.UUID

	qs.queueMutex.Lock()
	defer qs.queueMutex.Unlock()

	for category, genderQueues := range qs.queues {
		for gender, q := range genderQueues {
			for i := len(q.entries) - 1; i >= 0; i-- {
				entry := q.entries[i]
				if now.After(entry.ExpiresAt) || !qs.isUserConnected(entry.UserID) {
					expiredUsers = append(expiredUsers, entry.UserID)
					q.entries = append(q.entries[:i], q.entries[i+1:]...)
					close(entry.MatchChan)
				}
			}
			qs.updatePositionsForQueue(category, gender, 0)
		}
	}

	if len(expiredUsers) > 0 {
		qs.posMutex.Lock()
		qs.connMutex.Lock()
		for _, userID := range expiredUsers {
			delete(qs.userPositions, userID)
			delete(qs.userConnections, userID)
		}
		qs.connMutex.Unlock()
		qs.posMutex.Unlock()

		slog.Info("Cleaned up expired queue entries", "count", len(expiredUsers))
	}
}

func (qs *QueueService) Stop() {
	qs.queueMutex.Lock()
	defer qs.queueMutex.Unlock()

	for _, genderQueues := range qs.queues {
		for _, q := range genderQueues {
			for _, entry := range q.entries {
				close(entry.MatchChan)
			}
		}
	}
}
