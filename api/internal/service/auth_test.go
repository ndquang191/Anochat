package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAuthService(secret string) *AuthService {
	return &AuthService{
		jwtSecret:       []byte(secret),
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
