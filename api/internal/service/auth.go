package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
)

type AuthService struct {
	oauthConfig        *oauth2.Config
	jwtSecret          []byte
	userService        *UserService
	rdb                *redis.Client
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

func NewAuthService(
	userService *UserService,
	oauthConfig *oauth2.Config,
	jwtSecret string,
	rdb *redis.Client,
	accessTokenExpiry time.Duration,
	refreshTokenExpiry time.Duration,
) *AuthService {
	return &AuthService{
		oauthConfig:        oauthConfig,
		jwtSecret:          []byte(jwtSecret),
		userService:        userService,
		rdb:                rdb,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// OAuthResult holds both tokens returned after successful OAuth.
type OAuthResult struct {
	User         *identity.User
	AccessToken  string
	RefreshToken string
}

func (s *AuthService) ProcessOAuthCallback(ctx context.Context, code string) (*OAuthResult, error) {
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	googleUser, err := s.getUserInfoFromToken(ctx, token)
	if err != nil {
		return nil, err
	}

	slog.Info("Google OAuth user info received", "email", googleUser.Email, "name", googleUser.Name)

	user, err := s.userService.GetOrCreateUser(ctx, googleUser.Email, googleUser.Name, googleUser.Picture)
	if err != nil {
		return nil, err
	}

	// Auto-calculate age from Google birthday (best-effort, silently ignored if unavailable)
	if age := s.getAgeFromGoogle(ctx, token); age != nil {
		s.userService.UpdateProfile(ctx, user.ID, nil, nil, age, nil)
	}

	if user.Email == nil {
		return nil, fmt.Errorf("user email is nil")
	}

	accessToken, err := s.generateJWT(user.ID, *user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &OAuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

type googleBirthdayResponse struct {
	Birthdays []struct {
		Date struct {
			Year  int `json:"year"`
			Month int `json:"month"`
			Day   int `json:"day"`
		} `json:"date"`
	} `json:"birthdays"`
}

func (s *AuthService) getAgeFromGoogle(ctx context.Context, token *oauth2.Token) *int {
	client := s.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://people.googleapis.com/v1/people/me?personFields=birthdays")
	if err != nil {
		slog.Debug("Failed to fetch birthday from Google People API", "error", err)
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		slog.Debug("Google People API returned non-OK status", "status", resp.StatusCode)
		return nil
	}
	defer resp.Body.Close()

	var data googleBirthdayResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}

	for _, b := range data.Birthdays {
		if b.Date.Year <= 0 {
			continue
		}
		now := time.Now()
		age := now.Year() - b.Date.Year
		if b.Date.Month > 0 && b.Date.Day > 0 {
			if now.Before(time.Date(now.Year(), time.Month(b.Date.Month), b.Date.Day, 0, 0, 0, 0, time.UTC)) {
				age--
			}
		}
		return &age
	}
	return nil
}

func (s *AuthService) getUserInfoFromToken(ctx context.Context, token *oauth2.Token) (*identity.GoogleUserInfo, error) {
	client := s.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var userInfo identity.GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}
	return &userInfo, nil
}

func (s *AuthService) generateJWT(userID uuid.UUID, email string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(s.accessTokenExpiry)

	claims := identity.JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "anochat-api",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}
	return tokenString, nil
}

func (s *AuthService) ValidateJWT(tokenString string) (*identity.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &identity.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	if claims, ok := token.Claims.(*identity.JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, apperr.ErrInvalidJWT
}

// --- Refresh Token ---

const refreshTokenPrefix = "refresh:"

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *AuthService) generateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(b)

	key := refreshTokenPrefix + hashToken(token)
	if err := s.rdb.Set(ctx, key, userID.String(), s.refreshTokenExpiry).Err(); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return token, nil
}

// RefreshAccessToken validates a refresh token and returns a new access token.
// Uses a sliding window: the refresh token TTL is reset on each use.
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	key := refreshTokenPrefix + hashToken(refreshToken)
	userIDStr, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", apperr.ErrInvalidRefreshToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", apperr.ErrCorruptTokenData
	}

	user, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return "", apperr.ErrUserNotFound
	}
	if !user.IsActive {
		if err := s.rdb.Del(ctx, key).Err(); err != nil {
			slog.Warn("Failed to revoke refresh token for suspended user", "error", err, "user_id", userID)
		}
		return "", apperr.ErrUserSuspended
	}

	if user.Email == nil {
		return "", fmt.Errorf("user email is nil")
	}

	// Sliding window: reset TTL so active users never get kicked out
	if err := s.rdb.Expire(ctx, key, s.refreshTokenExpiry).Err(); err != nil {
		slog.Warn("Failed to extend refresh token TTL", "error", err)
	}

	return s.generateJWT(userID, *user.Email)
}

// RevokeRefreshToken removes a refresh token from Redis (used on logout).
func (s *AuthService) RevokeRefreshToken(ctx context.Context, refreshToken string) {
	key := refreshTokenPrefix + hashToken(refreshToken)
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		slog.Warn("Failed to revoke refresh token", "error", err)
	}
}
