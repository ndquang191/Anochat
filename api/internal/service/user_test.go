package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newUserServiceWithMocks() (*UserService, *mockUserRepo, *mockProfileRepo) {
	userRepo := new(mockUserRepo)
	profileRepo := new(mockProfileRepo)
	return NewUserService(userRepo, profileRepo), userRepo, profileRepo
}

func TestCreateUser(t *testing.T) {
	svc, userRepo, _ := newUserServiceWithMocks()

	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*identity.User")).Return(nil)

	user, err := svc.CreateUser(context.Background(), "test@test.com", "Test", "https://avatar.url")
	require.NoError(t, err)
	assert.Equal(t, "test@test.com", *user.Email)
	assert.Equal(t, "Test", *user.Name)
	assert.True(t, user.IsActive)
	userRepo.AssertExpectations(t)
}

func TestGetUserByID(t *testing.T) {
	svc, userRepo, _ := newUserServiceWithMocks()
	userID := uuid.New()
	email := "test@test.com"
	expected := &identity.User{ID: userID, Email: &email}

	userRepo.On("FindByID", mock.Anything, userID).Return(expected, nil)

	user, err := svc.GetUserByID(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, userID, user.ID)
	userRepo.AssertExpectations(t)
}

func TestGetUserByID_NotFound(t *testing.T) {
	svc, userRepo, _ := newUserServiceWithMocks()
	userID := uuid.New()

	userRepo.On("FindByID", mock.Anything, userID).Return(nil, repository.ErrNotFound)

	user, err := svc.GetUserByID(context.Background(), userID)
	assert.ErrorIs(t, err, apperr.ErrUserNotFound)
	assert.Nil(t, user)
	userRepo.AssertExpectations(t)
}

func TestGetUserWithProfile(t *testing.T) {
	svc, userRepo, _ := newUserServiceWithMocks()
	userID := uuid.New()
	age := 25
	expected := &identity.User{
		ID: userID,
		Profile: &identity.Profile{
			UserID: userID,
			Age:    &age,
		},
	}

	userRepo.On("FindByIDWithProfile", mock.Anything, userID).Return(expected, nil)

	user, err := svc.GetUserWithProfile(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, userID, user.ID)
	assert.NotNil(t, user.Profile)
	assert.Equal(t, 25, *user.Profile.Age)
	userRepo.AssertExpectations(t)
}

func TestGetOrCreateUser_Existing(t *testing.T) {
	svc, userRepo, _ := newUserServiceWithMocks()
	email := "existing@test.com"
	existing := &identity.User{ID: uuid.New(), Email: &email}

	userRepo.On("FindByEmail", mock.Anything, email).Return(existing, nil)

	user, err := svc.GetOrCreateUser(context.Background(), email, "Name", "https://avatar.url")
	require.NoError(t, err)
	assert.Equal(t, existing.ID, user.ID)
	// Should NOT call Create
	userRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestGetOrCreateUser_New(t *testing.T) {
	svc, userRepo, _ := newUserServiceWithMocks()
	email := "new@test.com"

	userRepo.On("FindByEmail", mock.Anything, email).Return(nil, repository.ErrNotFound)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*identity.User")).Return(nil)

	user, err := svc.GetOrCreateUser(context.Background(), email, "New User", "https://avatar.url")
	require.NoError(t, err)
	assert.Equal(t, email, *user.Email)
	userRepo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestGetProfile_Existing(t *testing.T) {
	svc, _, profileRepo := newUserServiceWithMocks()
	userID := uuid.New()
	age := 20
	expected := &identity.Profile{UserID: userID, Age: &age}

	profileRepo.On("FindByUserID", mock.Anything, userID).Return(expected, nil)

	profile, err := svc.GetProfile(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, 20, *profile.Age)
	profileRepo.AssertExpectations(t)
}

func TestGetProfile_AutoCreate(t *testing.T) {
	svc, _, profileRepo := newUserServiceWithMocks()
	userID := uuid.New()

	profileRepo.On("FindByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)
	profileRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *identity.Profile) bool {
		return p.UserID == userID && !p.IsHidden
	})).Return(nil)

	profile, err := svc.GetProfile(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, userID, profile.UserID)
	assert.False(t, profile.IsHidden)
	profileRepo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestUpdateProfile(t *testing.T) {
	svc, _, profileRepo := newUserServiceWithMocks()
	userID := uuid.New()
	existing := &identity.Profile{
		UserID:   userID,
		IsHidden: false,
	}

	profileRepo.On("FindByUserID", mock.Anything, userID).Return(existing, nil)
	profileRepo.On("Update", mock.Anything, mock.MatchedBy(func(p *identity.Profile) bool {
		return p.Nickname != nil && *p.Nickname == "newname" &&
			p.IsMale != nil && *p.IsMale == true &&
			p.Age != nil && *p.Age == 25 &&
			p.IsHidden == true
	})).Return(nil)

	nickname := "newname"
	isMale := true
	age := 25
	isHidden := true

	profile, err := svc.UpdateProfile(context.Background(), userID, &nickname, &isMale, &age, &isHidden)
	require.NoError(t, err)
	assert.Equal(t, "newname", *profile.Nickname)
	assert.True(t, *profile.IsMale)
	assert.Equal(t, 25, *profile.Age)
	assert.True(t, profile.IsHidden)
	assert.False(t, profile.UpdatedAt.IsZero())

	_ = time.Now() // used import
	profileRepo.AssertExpectations(t)
}

func TestUpdateProfile_ClearNickname(t *testing.T) {
	svc, _, profileRepo := newUserServiceWithMocks()
	userID := uuid.New()
	oldNick := "oldname"
	existing := &identity.Profile{UserID: userID, Nickname: &oldNick}

	profileRepo.On("FindByUserID", mock.Anything, userID).Return(existing, nil)
	profileRepo.On("Update", mock.Anything, mock.MatchedBy(func(p *identity.Profile) bool {
		return p.Nickname == nil
	})).Return(nil)

	empty := ""
	_, err := svc.UpdateProfile(context.Background(), userID, &empty, nil, nil, nil)
	require.NoError(t, err)
	profileRepo.AssertExpectations(t)
}
