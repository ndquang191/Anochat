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

// MatchNotifier is an interface for notifying when matches are found
type MatchNotifier interface {
	NotifyMatch(user1ID, user2ID, roomID uuid.UUID, category string)
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

	// Match notifier for WebSocket
	matchNotifier MatchNotifier
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

	slog.Info("QueueService initialized successfully")

	// Start cleanup goroutine
	go qs.startCleanupRoutine()

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
		slog.Warn("User cannot join queue - already in queue", "user_id", userID, "category", category)
		// Log current queue status even when user cannot join
		qs.logCurrentQueueStatus(category)
		return nil, fmt.Errorf("user already in queue")
	}

	// Check if user has active room
	activeRoom, err := qs.userService.GetActiveRoom(ctx, userID)
	if err == nil && activeRoom != nil {
		slog.Warn("User cannot join queue - already has active room", "user_id", userID, "room_id", activeRoom.ID)
		return nil, fmt.Errorf("user already has active room")
	}

	// Create queue entry
	now := time.Now()
	entry := &QueueEntry{
		UserID:    userID,
		Profile:   profile,
		Category:  category,
		JoinedAt:  now,
		ExpiresAt: now.Add(qs.config.Chat.QueueHeartbeatTTL), // Set TTL expiration
		IsMatched: false,
		MatchChan: make(chan *MatchResult, 1),
	}

	// Add to appropriate queue and track position
	qs.queueMutex.Lock()
	isMale := profile.IsMale != nil && *profile.IsMale

	if isMale {
		// Initialize map if not exists
		if qs.queueMale[category] == nil {
			qs.queueMale[category] = make([]*QueueEntry, 0)
		}
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
		// Initialize map if not exists
		if qs.queueFemale[category] == nil {
			qs.queueFemale[category] = make([]*QueueEntry, 0)
		}
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

	// Log current queue status
	qs.logCurrentQueueStatus(category)

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
	isUnknown := entry.Profile.IsMale == nil

	// Strategy: Try opposite gender first (preferred), then same gender (fallback)
	var match *QueueEntry
	var matchedFromOppositeGender bool

	// 1. Priority: Try opposite gender first
	var oppositeQueue []*QueueEntry
	if isMale {
		oppositeQueue = qs.queueFemale[category]
	} else if !isUnknown {
		oppositeQueue = qs.queueMale[category]
	} else {
		// Unknown gender: try male queue first, then female
		oppositeQueue = append(qs.queueMale[category], qs.queueFemale[category]...)
	}

	for _, otherEntry := range oppositeQueue {
		if !otherEntry.IsMatched && qs.isUserConnected(otherEntry.UserID) && otherEntry.UserID != entry.UserID {
			match = otherEntry
			matchedFromOppositeGender = true
			break
		}
	}

	// 2. Fallback: Try same gender if no opposite gender match found
	if match == nil {
		var sameQueue []*QueueEntry
		if isMale {
			sameQueue = qs.queueMale[category]
		} else if !isUnknown {
			sameQueue = qs.queueFemale[category]
		} else {
			// Unknown already tried both queues above
			sameQueue = []*QueueEntry{}
		}

		for _, otherEntry := range sameQueue {
			if !otherEntry.IsMatched && qs.isUserConnected(otherEntry.UserID) && otherEntry.UserID != entry.UserID {
				match = otherEntry
				matchedFromOppositeGender = false
				break
			}
		}
	}

	// Debug: Log match attempt
	slog.Info("Match attempt",
		"user_id", entry.UserID,
		"category", category,
		"is_male", isMale,
		"is_unknown", isUnknown,
		"found_match", match != nil,
		"opposite_gender", matchedFromOppositeGender,
		"match_user_id", func() string {
			if match != nil {
				return match.UserID.String()
			}
			return "none"
		}())

	// If no match found, return
	if match == nil {
		return
	}

	// Mark both entries as matched
	entry.IsMatched = true
	match.IsMatched = true

	// Determine match's gender for queue removal
	matchIsMale := match.Profile.IsMale != nil && *match.Profile.IsMale

	// Remove both entries from their respective queues
	// Unknown gender users are stored in female queue (isMale=false)
	qs.removeEntryFromQueue(category, entry, isMale)
	qs.removeEntryFromQueue(category, match, matchIsMale)

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

	// Notify via WebSocket if notifier is set
	if qs.matchNotifier != nil {
		qs.matchNotifier.NotifyMatch(entry.UserID, match.UserID, room.ID, category)
	}

	// Log updated queue status after match
	qs.logCurrentQueueStatus(category)
}

// SetMatchNotifier sets the match notifier for WebSocket notifications
func (qs *QueueService) SetMatchNotifier(notifier MatchNotifier) {
	qs.matchNotifier = notifier
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

			// Remove from tracking maps
			qs.posMutex.Lock()
			delete(qs.userPositions, entry.UserID)
			qs.posMutex.Unlock()

			qs.connMutex.Lock()
			delete(qs.userConnections, entry.UserID)
			qs.connMutex.Unlock()

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
				// Refresh TTL as heartbeat
				entry.ExpiresAt = time.Now().Add(qs.config.Chat.QueueHeartbeatTTL)

				return &QueueStatus{
					IsInQueue: true,
					Position:  position.Position,
					Category:  category,
					JoinedAt:  position.JoinedAt,
					ExpiresAt: entry.ExpiresAt,
				}, nil
			}
		}
	}

	for category, entries := range qs.queueFemale {
		for _, entry := range entries {
			if entry.UserID == userID {
				// Refresh TTL as heartbeat
				entry.ExpiresAt = time.Now().Add(qs.config.Chat.QueueHeartbeatTTL)

				return &QueueStatus{
					IsInQueue: true,
					Position:  position.Position,
					Category:  category,
					JoinedAt:  position.JoinedAt,
					ExpiresAt: entry.ExpiresAt,
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
				// Log updated queue status
				qs.logCurrentQueueStatus(category)
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
				// Log updated queue status
				qs.logCurrentQueueStatus(category)
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

// logCurrentQueueStatus logs detailed information about who is currently in the queue
func (qs *QueueService) logCurrentQueueStatus(category string) {
	qs.queueMutex.RLock()
	defer qs.queueMutex.RUnlock()

	maleEntries := qs.queueMale[category]
	femaleEntries := qs.queueFemale[category]

	// Log male users in queue
	if len(maleEntries) > 0 {
		slog.Info("Male users in queue",
			"category", category,
			"count", len(maleEntries),
			"users", func() []string {
				users := make([]string, len(maleEntries))
				for i, entry := range maleEntries {
					users[i] = entry.UserID.String()
				}
				return users
			}())
	} else {
		slog.Info("No male users in queue", "category", category)
	}

	// Log female users in queue
	if len(femaleEntries) > 0 {
		slog.Info("Female users in queue",
			"category", category,
			"count", len(femaleEntries),
			"users", func() []string {
				users := make([]string, len(femaleEntries))
				for i, entry := range femaleEntries {
					users[i] = entry.UserID.String()
				}
				return users
			}())
	} else {
		slog.Info("No female users in queue", "category", category)
	}

	// Log summary
	slog.Info("Queue summary",
		"category", category,
		"male_count", len(maleEntries),
		"female_count", len(femaleEntries),
		"total_count", len(maleEntries)+len(femaleEntries))
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
				// Log updated queue status
				qs.logCurrentQueueStatus(category)
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
				// Log updated queue status
				qs.logCurrentQueueStatus(category)
				return nil
			}
		}
	}

	return fmt.Errorf("user not found in queue")
}

// startCleanupRoutine runs periodically to remove expired queue entries
func (qs *QueueService) startCleanupRoutine() {
	ticker := time.NewTicker(qs.config.Chat.QueueCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		qs.cleanupExpiredEntries()
	}
}

// cleanupExpiredEntries removes users who haven't sent heartbeat within TTL
func (qs *QueueService) cleanupExpiredEntries() {
	now := time.Now()
	expiredUsers := make([]uuid.UUID, 0)

	qs.queueMutex.Lock()
	defer qs.queueMutex.Unlock()

	// Check male queue
	for category, entries := range qs.queueMale {
		for i := len(entries) - 1; i >= 0; i-- {
			entry := entries[i]
			if now.After(entry.ExpiresAt) || !qs.isUserConnected(entry.UserID) {
				expiredUsers = append(expiredUsers, entry.UserID)
				// Remove from queue
				qs.queueMale[category] = append(entries[:i], entries[i+1:]...)
				close(entry.MatchChan)
			}
		}
	}

	// Check female queue
	for category, entries := range qs.queueFemale {
		for i := len(entries) - 1; i >= 0; i-- {
			entry := entries[i]
			if now.After(entry.ExpiresAt) || !qs.isUserConnected(entry.UserID) {
				expiredUsers = append(expiredUsers, entry.UserID)
				// Remove from queue
				qs.queueFemale[category] = append(entries[:i], entries[i+1:]...)
				close(entry.MatchChan)
			}
		}
	}

	// Clean up expired users from tracking maps
	if len(expiredUsers) > 0 {
		qs.posMutex.Lock()
		qs.connMutex.Lock()

		for _, userID := range expiredUsers {
			delete(qs.userPositions, userID)
			delete(qs.userConnections, userID)
		}

		qs.connMutex.Unlock()
		qs.posMutex.Unlock()

		// Update positions for remaining users
		for category := range qs.queueMale {
			qs.updatePositionsAfterRemoval(category, true, 0)
		}
		for category := range qs.queueFemale {
			qs.updatePositionsAfterRemoval(category, false, 0)
		}

		slog.Info("Cleaned up expired queue entries", "count", len(expiredUsers), "users", expiredUsers)
	}
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
