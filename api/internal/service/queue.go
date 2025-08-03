package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/model"
	"github.com/ndquang191/Anochat/api/pkg/config"
	"gorm.io/gorm"
)

// QueueEntry represents a user waiting in the queue
type QueueEntry struct {
	UserID    uuid.UUID
	Profile   *model.Profile
	Category  string
	JoinedAt  time.Time
	ExpiresAt time.Time
	IsMatched bool
	MatchChan chan *MatchResult
}

// MatchResult represents the result of a successful match
type MatchResult struct {
	RoomID   uuid.UUID
	User1ID  uuid.UUID
	User2ID  uuid.UUID
	Category string
	Error    error
}

// QueueStatus represents the status of a user in the queue
type QueueStatus struct {
	IsInQueue bool      `json:"is_in_queue"`
	Position  int       `json:"position"`
	Category  string    `json:"category"`
	JoinedAt  time.Time `json:"joined_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// QueuePosition tracks user position in queue
type QueuePosition struct {
	UserID   uuid.UUID
	Position int
	JoinedAt time.Time
}

// QueueService with enhanced features
type QueueService struct {
	db          *gorm.DB
	roomService *RoomService
	userService *UserService
	config      *config.Config

	// Separate queues for male and female
	queueMale   map[string][]*QueueEntry // category -> male entries
	queueFemale map[string][]*QueueEntry // category -> female entries
	queueMutex  sync.RWMutex

	// User connection tracking
	userConnections map[uuid.UUID]bool // userID -> isConnected
	connMutex       sync.RWMutex

	// Position tracking for quick lookup
	userPositions map[uuid.UUID]*QueuePosition // userID -> position info
	posMutex      sync.RWMutex

	// Match statistics
	matchStats struct {
		sync.RWMutex
		totalMatches   int64
		maleWaitTime   time.Duration
		femaleWaitTime time.Duration
		lastMatchTime  time.Time
	}
}

// NewQueueService creates a new enhanced queue service
func NewQueueService(db *gorm.DB, roomService *RoomService, userService *UserService, config *config.Config) *QueueService {
	qs := &QueueService{
		db:              db,
		roomService:     roomService,
		userService:     userService,
		config:          config,
		queueMale:       make(map[string][]*QueueEntry),
		queueFemale:     make(map[string][]*QueueEntry),
		userConnections: make(map[uuid.UUID]bool),
		userPositions:   make(map[uuid.UUID]*QueuePosition),
	}

	return qs
}

// JoinQueue adds a user to the appropriate queue based on gender
func (qs *QueueService) JoinQueue(ctx context.Context, userID uuid.UUID, category string) (*QueueEntry, error) {
	// Validate category
	if !config.IsValidCategory(category) {
		return nil, fmt.Errorf("invalid category: %s", category)
	}

	// Get user profile
	profile, err := qs.userService.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	if profile == nil {
		return nil, fmt.Errorf("user profile not found")
	}

	// Check if user is already in queue
	if qs.isUserInQueue(userID) {
		return nil, fmt.Errorf("user already in queue")
	}

	// Check if user has active room
	activeRoom, err := qs.userService.GetActiveRoom(ctx, userID)
	if err == nil && activeRoom != nil {
		return nil, fmt.Errorf("user already has active room")
	}

	// Create queue entry
	now := time.Now()
	entry := &QueueEntry{
		UserID:    userID,
		Profile:   profile,
		Category:  category,
		JoinedAt:  now,
		ExpiresAt: time.Time{}, // No expiration
		IsMatched: false,
		MatchChan: make(chan *MatchResult, 1),
	}

	// Add to appropriate queue and track position
	qs.queueMutex.Lock()
	isMale := profile.IsMale != nil && *profile.IsMale

	if isMale {
		position := len(qs.queueMale[category]) + 1
		qs.queueMale[category] = append(qs.queueMale[category], entry)

		// Track position
		qs.posMutex.Lock()
		qs.userPositions[userID] = &QueuePosition{
			UserID:   userID,
			Position: position,
			JoinedAt: now,
		}
		qs.posMutex.Unlock()
	} else {
		position := len(qs.queueFemale[category]) + 1
		qs.queueFemale[category] = append(qs.queueFemale[category], entry)

		// Track position
		qs.posMutex.Lock()
		qs.userPositions[userID] = &QueuePosition{
			UserID:   userID,
			Position: position,
			JoinedAt: now,
		}
		qs.posMutex.Unlock()
	}
	qs.queueMutex.Unlock()

	// Mark user as connected
	qs.connMutex.Lock()
	qs.userConnections[userID] = true
	qs.connMutex.Unlock()

	slog.Info("User joined queue",
		"user_id", userID,
		"category", category,
		"gender", qs.getGenderString(profile),
		"position", qs.userPositions[userID].Position,
		"joined_at", entry.JoinedAt)

	// Try to find a match immediately
	go qs.tryMatch(entry)

	return entry, nil
}

// tryMatch attempts to find a match using separate queues
func (qs *QueueService) tryMatch(entry *QueueEntry) {
	qs.queueMutex.Lock()
	defer qs.queueMutex.Unlock()

	// Check if entry is still valid and user is still connected
	if entry.IsMatched || !qs.isUserConnected(entry.UserID) {
		return
	}

	category := entry.Category
	isMale := entry.Profile.IsMale != nil && *entry.Profile.IsMale

	// Get the opposite gender queue
	var oppositeQueue []*QueueEntry
	if isMale {
		oppositeQueue = qs.queueFemale[category]
	} else {
		oppositeQueue = qs.queueMale[category]
	}

	// Find the first available match (FIFO) - only consider connected users
	var match *QueueEntry
	for _, otherEntry := range oppositeQueue {
		if !otherEntry.IsMatched && qs.isUserConnected(otherEntry.UserID) {
			match = otherEntry
			break
		}
	}

	// If no match found, return
	if match == nil {
		return
	}

	// Mark both entries as matched
	entry.IsMatched = true
	match.IsMatched = true

	// Remove both entries from their respective queues
	qs.removeEntryFromQueue(category, entry, isMale)
	qs.removeEntryFromQueue(category, match, !isMale)

	// Update match statistics
	qs.updateMatchStats(entry, match)

	// Create room
	room, err := qs.roomService.CreateRoom(context.Background(), entry.UserID, match.UserID, category)
	if err != nil {
		// Send error to both users
		entry.MatchChan <- &MatchResult{Error: err}
		match.MatchChan <- &MatchResult{Error: err}
		return
	}

	// Send success to both users
	result := &MatchResult{
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
		"category", category,
		"wait_time_1", time.Since(entry.JoinedAt),
		"wait_time_2", time.Since(match.JoinedAt))
}

// updateMatchStats updates match statistics
func (qs *QueueService) updateMatchStats(entry1, entry2 *QueueEntry) {
	qs.matchStats.Lock()
	defer qs.matchStats.Unlock()

	qs.matchStats.totalMatches++
	qs.matchStats.lastMatchTime = time.Now()

	// Calculate wait times
	waitTime1 := time.Since(entry1.JoinedAt)
	waitTime2 := time.Since(entry2.JoinedAt)

	if entry1.Profile.IsMale != nil && *entry1.Profile.IsMale {
		qs.matchStats.maleWaitTime = (qs.matchStats.maleWaitTime + waitTime1) / 2
		qs.matchStats.femaleWaitTime = (qs.matchStats.femaleWaitTime + waitTime2) / 2
	} else {
		qs.matchStats.maleWaitTime = (qs.matchStats.maleWaitTime + waitTime2) / 2
		qs.matchStats.femaleWaitTime = (qs.matchStats.femaleWaitTime + waitTime1) / 2
	}
}

// removeEntryFromQueue removes an entry and updates positions
func (qs *QueueService) removeEntryFromQueue(category string, entry *QueueEntry, isMale bool) {
	var queue []*QueueEntry
	if isMale {
		queue = qs.queueMale[category]
	} else {
		queue = qs.queueFemale[category]
	}

	for i, e := range queue {
		if e.UserID == entry.UserID {
			if isMale {
				qs.queueMale[category] = append(queue[:i], queue[i+1:]...)
			} else {
				qs.queueFemale[category] = append(queue[:i], queue[i+1:]...)
			}

			// Update positions for remaining users
			qs.updatePositionsAfterRemoval(category, isMale, i)
			break
		}
	}
}

// updatePositionsAfterRemoval updates positions after a user is removed
func (qs *QueueService) updatePositionsAfterRemoval(category string, isMale bool, removedIndex int) {
	var queue []*QueueEntry
	if isMale {
		queue = qs.queueMale[category]
	} else {
		queue = qs.queueFemale[category]
	}

	qs.posMutex.Lock()
	defer qs.posMutex.Unlock()

	// Update positions for users after the removed one
	for i, entry := range queue {
		if i >= removedIndex {
			if pos, exists := qs.userPositions[entry.UserID]; exists {
				pos.Position = i + 1
			}
		}
	}
}

// GetQueueStatus returns the current queue status for a user (O(1))
func (qs *QueueService) GetQueueStatus(ctx context.Context, userID uuid.UUID) (*QueueStatus, error) {
	qs.posMutex.RLock()
	position, exists := qs.userPositions[userID]
	qs.posMutex.RUnlock()

	if !exists {
		return &QueueStatus{
			IsInQueue: false,
			Position:  0,
			Category:  "",
			JoinedAt:  time.Time{},
			ExpiresAt: time.Time{},
		}, nil
	}

	// Find category by scanning queues (could be optimized further)
	qs.queueMutex.RLock()
	defer qs.queueMutex.RUnlock()

	for category, entries := range qs.queueMale {
		for _, entry := range entries {
			if entry.UserID == userID {
				return &QueueStatus{
					IsInQueue: true,
					Position:  position.Position,
					Category:  category,
					JoinedAt:  position.JoinedAt,
					ExpiresAt: time.Time{},
				}, nil
			}
		}
	}

	for category, entries := range qs.queueFemale {
		for _, entry := range entries {
			if entry.UserID == userID {
				return &QueueStatus{
					IsInQueue: true,
					Position:  position.Position,
					Category:  category,
					JoinedAt:  position.JoinedAt,
					ExpiresAt: time.Time{},
				}, nil
			}
		}
	}

	return &QueueStatus{IsInQueue: false}, nil
}

// GetMatchStatistics returns match statistics
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

// GetQueueStats returns detailed queue statistics
func (qs *QueueService) GetQueueStats() map[string]map[string]interface{} {
	qs.queueMutex.RLock()
	defer qs.queueMutex.RUnlock()

	stats := make(map[string]map[string]interface{})

	for category := range qs.queueMale {
		maleCount := len(qs.queueMale[category])
		femaleCount := len(qs.queueFemale[category])

		stats[category] = map[string]interface{}{
			"male_count":          maleCount,
			"female_count":        femaleCount,
			"total_count":         maleCount + femaleCount,
			"estimated_wait_time": qs.calculateEstimatedWaitTime(category, maleCount, femaleCount),
		}
	}

	return stats
}

// calculateEstimatedWaitTime estimates wait time based on queue balance
func (qs *QueueService) calculateEstimatedWaitTime(category string, maleCount, femaleCount int) time.Duration {
	// Simple estimation: if more people of opposite gender, faster match
	if maleCount == 0 || femaleCount == 0 {
		return 5 * time.Minute // No matches possible
	}

	// If balanced, estimate 1-2 minutes
	if abs(maleCount-femaleCount) <= 2 {
		return 90 * time.Second
	}

	// If unbalanced, longer wait
	return time.Duration(max(maleCount, femaleCount)) * 30 * time.Second
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// UserDisconnected marks a user as disconnected and removes them from queue
func (qs *QueueService) UserDisconnected(userID uuid.UUID) {
	qs.connMutex.Lock()
	qs.userConnections[userID] = false
	qs.connMutex.Unlock()

	// Remove from queue
	qs.queueMutex.Lock()
	defer qs.queueMutex.Unlock()

	// Remove from male queue
	for category, entries := range qs.queueMale {
		for i, entry := range entries {
			if entry.UserID == userID {
				qs.queueMale[category] = append(entries[:i], entries[i+1:]...)
				close(entry.MatchChan)

				// Update positions
				qs.updatePositionsAfterRemoval(category, true, i)

				// Remove from position tracking
				qs.posMutex.Lock()
				delete(qs.userPositions, userID)
				qs.posMutex.Unlock()

				slog.Info("User disconnected from male queue", "user_id", userID, "category", category)
				return
			}
		}
	}

	// Remove from female queue
	for category, entries := range qs.queueFemale {
		for i, entry := range entries {
			if entry.UserID == userID {
				qs.queueFemale[category] = append(entries[:i], entries[i+1:]...)
				close(entry.MatchChan)

				// Update positions
				qs.updatePositionsAfterRemoval(category, false, i)

				// Remove from position tracking
				qs.posMutex.Lock()
				delete(qs.userPositions, userID)
				qs.posMutex.Unlock()

				slog.Info("User disconnected from female queue", "user_id", userID, "category", category)
				return
			}
		}
	}
}

// isUserConnected checks if a user is still connected
func (qs *QueueService) isUserConnected(userID uuid.UUID) bool {
	qs.connMutex.RLock()
	defer qs.connMutex.RUnlock()
	return qs.userConnections[userID]
}

// isUserInQueue checks if a user is already in any queue
func (qs *QueueService) isUserInQueue(userID uuid.UUID) bool {
	qs.posMutex.RLock()
	defer qs.posMutex.RUnlock()
	_, exists := qs.userPositions[userID]
	return exists
}

// getGenderString helper function
func (qs *QueueService) getGenderString(profile *model.Profile) string {
	if profile.IsMale == nil {
		return "unknown"
	}
	if *profile.IsMale {
		return "male"
	}
	return "female"
}

// LeaveQueue removes a user from the queue (manual leave)
func (qs *QueueService) LeaveQueue(ctx context.Context, userID uuid.UUID) error {
	qs.queueMutex.Lock()
	defer qs.queueMutex.Unlock()

	// Remove from male queue
	for category, entries := range qs.queueMale {
		for i, entry := range entries {
			if entry.UserID == userID {
				qs.queueMale[category] = append(entries[:i], entries[i+1:]...)
				close(entry.MatchChan)

				// Update positions
				qs.updatePositionsAfterRemoval(category, true, i)

				// Mark as disconnected
				qs.connMutex.Lock()
				qs.userConnections[userID] = false
				qs.connMutex.Unlock()

				// Remove from position tracking
				qs.posMutex.Lock()
				delete(qs.userPositions, userID)
				qs.posMutex.Unlock()

				slog.Info("User manually left male queue", "user_id", userID, "category", category)
				return nil
			}
		}
	}

	// Remove from female queue
	for category, entries := range qs.queueFemale {
		for i, entry := range entries {
			if entry.UserID == userID {
				qs.queueFemale[category] = append(entries[:i], entries[i+1:]...)
				close(entry.MatchChan)

				// Update positions
				qs.updatePositionsAfterRemoval(category, false, i)

				// Mark as disconnected
				qs.connMutex.Lock()
				qs.userConnections[userID] = false
				qs.connMutex.Unlock()

				// Remove from position tracking
				qs.posMutex.Lock()
				delete(qs.userPositions, userID)
				qs.posMutex.Unlock()

				slog.Info("User manually left female queue", "user_id", userID, "category", category)
				return nil
			}
		}
	}

	return fmt.Errorf("user not found in queue")
}

// Stop stops the queue service
func (qs *QueueService) Stop() {
	qs.queueMutex.Lock()
	defer qs.queueMutex.Unlock()

	// Close all match channels in male queue
	for _, entries := range qs.queueMale {
		for _, entry := range entries {
			close(entry.MatchChan)
		}
	}

	// Close all match channels in female queue
	for _, entries := range qs.queueFemale {
		for _, entry := range entries {
			close(entry.MatchChan)
		}
	}
}
