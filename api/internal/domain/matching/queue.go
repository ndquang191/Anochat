package matching

import "github.com/google/uuid"

// MatchNotifier is an interface for notifying when matches are found.
type MatchNotifier interface {
	NotifyMatch(user1ID, user2ID, roomID uuid.UUID)
	NotifyFakeMatch(session *FakeSession)
	NotifyFakePartnerLeft(userID, roomID uuid.UUID)
}
