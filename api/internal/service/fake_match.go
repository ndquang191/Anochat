package service

import (
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/internal/domain/matching"
)

var (
	fakeMaleNames = []string{
		"Minh", "Huy", "Nam", "Đức", "Khang", "Phúc", "Long", "Khôi", "Quân", "Tuấn",
	}
	fakeFemaleNames = []string{
		"Linh", "An", "Vy", "Nhi", "Trang", "Hân", "Ngọc", "Mai", "Thảo", "Trâm",
	}
)

const (
	fakeReplyTimeout = 40 * time.Second
	fakeLeaveDelay   = 1500 * time.Millisecond
)

type FakeMatchService struct {
	mu             sync.RWMutex
	sessionsByUser map[uuid.UUID]*matching.FakeSession
	sessionsByRoom map[uuid.UUID]*matching.FakeSession
	rng            *rand.Rand
}

func NewFakeMatchService() *FakeMatchService {
	return &FakeMatchService{
		sessionsByUser: make(map[uuid.UUID]*matching.FakeSession),
		sessionsByRoom: make(map[uuid.UUID]*matching.FakeSession),
		rng:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *FakeMatchService) CreateSession(userID uuid.UUID, profile *identity.Profile) *matching.FakeSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing := s.sessionsByUser[userID]; existing != nil {
		return existing
	}

	partnerIsMale := s.randomPartnerGender(profile)
	createdAt := time.Now().UTC()
	session := &matching.FakeSession{
		UserID:        userID,
		RoomID:        uuid.New(),
		PartnerID:     uuid.New(),
		GreetingID:    uuid.New(),
		PartnerName:   s.randomName(partnerIsMale),
		PartnerIsMale: partnerIsMale,
		PartnerAge:    18 + s.rng.Intn(13),
		Greeting:      fakeGreetingMessages[s.rng.Intn(len(fakeGreetingMessages))],
		PromptOrder:   s.randomPromptOrder(),
		AwaitingReply: true,
		CreatedAt:     createdAt,
		EndsAt:        createdAt.Add(fakeReplyTimeout),
	}

	s.sessionsByUser[userID] = session
	s.sessionsByRoom[session.RoomID] = session
	return session
}

func (s *FakeMatchService) GetByUserID(userID uuid.UUID) *matching.FakeSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionsByUser[userID]
}

func (s *FakeMatchService) GetByRoomID(roomID uuid.UUID) *matching.FakeSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionsByRoom[roomID]
}

func (s *FakeMatchService) HasActiveSession(userID uuid.UUID) bool {
	return s.GetByUserID(userID) != nil
}

func (s *FakeMatchService) EndSession(userID uuid.UUID) *matching.FakeSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := s.sessionsByUser[userID]
	if session == nil {
		return nil
	}

	delete(s.sessionsByUser, userID)
	delete(s.sessionsByRoom, session.RoomID)
	return session
}

func (s *FakeMatchService) IsParticipant(userID, roomID uuid.UUID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session := s.sessionsByRoom[roomID]
	return session != nil && session.UserID == userID
}

func (s *FakeMatchService) BuildRoomDTO(session *matching.FakeSession) *chat.Room {
	if session == nil {
		return nil
	}

	return &chat.Room{
		ID:        session.RoomID,
		User1ID:   session.UserID,
		User2ID:   session.PartnerID,
		CreatedAt: session.CreatedAt,
	}
}

func (s *FakeMatchService) BuildGreetingMessage(session *matching.FakeSession) *chat.Message {
	if session == nil {
		return nil
	}

	return &chat.Message{
		ID:        session.GreetingID,
		RoomID:    session.RoomID,
		SenderID:  session.PartnerID,
		Content:   session.Greeting,
		CreatedAt: session.CreatedAt,
	}
}

func (s *FakeMatchService) AdvanceConversation(userID, roomID uuid.UUID) *chat.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := s.sessionsByUser[userID]
	if session == nil || session.RoomID != roomID || !session.AwaitingReply {
		return nil
	}

	now := time.Now().UTC()

	if session.PromptIndex >= len(session.PromptOrder) {
		session.AwaitingReply = false
		session.EndsAt = now.Add(fakeLeaveDelay)
		return nil
	}

	topic := fakeChatTopic(session.PromptOrder[session.PromptIndex])
	session.PromptIndex++
	session.AwaitingReply = true
	session.EndsAt = now.Add(fakeReplyTimeout)

	variants := fakePromptVariants[topic]
	if len(variants) == 0 {
		return nil
	}

	return &chat.Message{
		ID:        uuid.New(),
		RoomID:    session.RoomID,
		SenderID:  session.PartnerID,
		Content:   variants[s.rng.Intn(len(variants))],
		CreatedAt: now,
	}
}

func (s *FakeMatchService) ShouldEndSession(userID, roomID uuid.UUID, now time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session := s.sessionsByUser[userID]
	if session == nil || session.RoomID != roomID {
		return false
	}

	return !session.EndsAt.IsZero() && !now.Before(session.EndsAt)
}

func (s *FakeMatchService) randomPartnerGender(profile *identity.Profile) bool {
	if profile != nil && profile.IsMale != nil {
		return !*profile.IsMale
	}
	return s.rng.Intn(2) == 0
}

func (s *FakeMatchService) randomName(isMale bool) string {
	if isMale {
		return fakeMaleNames[s.rng.Intn(len(fakeMaleNames))]
	}
	return fakeFemaleNames[s.rng.Intn(len(fakeFemaleNames))]
}

func (s *FakeMatchService) randomPromptOrder() []string {
	topics := []string{
		string(fakeTopicName),
		string(fakeTopicAge),
		string(fakeTopicLocation),
	}

	s.rng.Shuffle(len(topics), func(i, j int) {
		topics[i], topics[j] = topics[j], topics[i]
	})

	return topics
}
