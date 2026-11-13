package service

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestAuthService(secret string) *AuthService {
	return &AuthService{
		jwtSecret:         []byte(secret),
		accessTokenExpiry: 15 * time.Minute,
	}
}

func TestGenerateAndValidateJWT(t *testing.T) {
	svc := newTestAuthService("test-secret-key-at-least-32-bytes!")
	userID := uuid.New()
	email := "test@example.com"

	token, err := svc.generateJWT(userID, email)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := svc.ValidateJWT(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, "anochat-api", claims.Issuer)
	assert.Equal(t, userID.String(), claims.Subject)
}

func TestValidateJWT_Expired(t *testing.T) {
	svc := newTestAuthService("test-secret-key-at-least-32-bytes!")
	userID := uuid.New()

	// Create a token that's already expired
	now := time.Now().Add(-25 * time.Hour) // 25 hours ago
	claims := identity.JWTClaims{
		UserID: userID,
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)), // expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "anochat-api",
			Subject:   userID.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(svc.jwtSecret)
	require.NoError(t, err)

	_, err = svc.ValidateJWT(tokenString)
	assert.Error(t, err)
}

func TestValidateJWT_InvalidSignature(t *testing.T) {
	svc1 := newTestAuthService("secret-one-at-least-32-bytes!!!!")
	svc2 := newTestAuthService("secret-two-at-least-32-bytes!!!!")

	token, err := svc1.generateJWT(uuid.New(), "test@example.com")
	require.NoError(t, err)

	_, err = svc2.ValidateJWT(token)
	assert.Error(t, err)
}

func TestValidateJWT_Malformed(t *testing.T) {
	svc := newTestAuthService("test-secret-key-at-least-32-bytes!")

	_, err := svc.ValidateJWT("not-a-valid-jwt")
	assert.Error(t, err)

	_, err = svc.ValidateJWT("")
	assert.Error(t, err)
}

func TestRefreshTokenRotatesAndConcurrentRetryConverges(t *testing.T) {
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	userService, userRepo, _ := newUserServiceWithMocks()
	userID := uuid.New()
	email := "person@example.com"
	userRepo.On("FindByID", mock.Anything, userID).
		Return(&identity.User{ID: userID, Email: &email}, nil)

	svc := &AuthService{
		jwtSecret:          []byte("test-secret-key-at-least-32-bytes!"),
		userService:        userService,
		rdb:                redisClient,
		accessTokenExpiry:  15 * time.Minute,
		refreshTokenExpiry: 30 * 24 * time.Hour,
	}
	oldToken, err := svc.generateRefreshToken(context.Background(), userID)
	require.NoError(t, err)

	first, err := svc.RefreshAccessToken(context.Background(), oldToken)
	require.NoError(t, err)
	require.NotEqual(t, oldToken, first.RefreshToken)
	require.NotEmpty(t, first.AccessToken)

	second, err := svc.RefreshAccessToken(context.Background(), oldToken)
	require.NoError(t, err)
	assert.Equal(t, first.RefreshToken, second.RefreshToken)

	svc.RevokeRefreshToken(context.Background(), oldToken)
	_, err = svc.RefreshAccessToken(context.Background(), first.RefreshToken)
	assert.ErrorIs(t, err, apperr.ErrInvalidRefreshToken)
}
