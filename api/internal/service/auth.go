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
	"strings"
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

// CreateDevSession creates a normal application session without contacting an
// OAuth provider. The caller is responsible for exposing this only in local
// development.
func (s *AuthService) CreateDevSession(ctx context.Context, email, name string) (*OAuthResult, error) {
	user, err := s.userService.GetOrCreateUser(ctx, email, name, "")
	if err != nil {
		return nil, err
	}

	accessToken, err := s.generateJWT(user.ID, email)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &OAuthResult{User: user, AccessToken: accessToken, RefreshToken: refreshToken}, nil
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

	if !googleUser.Verified || googleUser.ID == "" || googleUser.Email == "" {
		return nil, fmt.Errorf("Google account email is missing or unverified")
	}
	slog.Info("Google OAuth user info received", "google_subject", googleUser.ID)

	user, err := s.userService.GetOrCreateOAuthUser(
		ctx, "google", googleUser.ID, googleUser.Email, googleUser.Name, googleUser.Picture,
	)
	if err != nil {
		return nil, err
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
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer("anochat-api"), jwt.WithExpirationRequired())
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
const refreshRotationGrace = 30 * time.Second

var rotateRefreshTokenScript = redis.NewScript(`
local current = redis.call("GET", KEYS[1])
if not current then
    return {0, ""}
end
if string.sub(current, 1, 8) == "rotated:" then
    return {2, string.sub(current, 9)}
end
local user_id = current
if string.sub(current, 1, 5) == "user:" then
    user_id = string.sub(current, 6)
end
redis.call("SET", KEYS[2], "user:" .. user_id, "PX", ARGV[2])
redis.call("SET", KEYS[1], "rotated:" .. ARGV[1], "PX", ARGV[3])
return {1, user_id}
`)

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *AuthService) generateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	token, err := newRefreshToken()
	if err != nil {
		return "", err
	}

	key := refreshTokenPrefix + hashToken(token)
	if err := s.rdb.Set(ctx, key, "user:"+userID.String(), s.refreshTokenExpiry).Err(); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return token, nil
}

func newRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

type TokenRefreshResult struct {
	AccessToken  string
	RefreshToken string
}

// RefreshAccessToken rotates the refresh token atomically. The old token keeps
// a short mapping to the replacement so concurrent browser tabs converge on
// the same new token instead of invalidating one another.
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenRefreshResult, error) {
	newToken, err := newRefreshToken()
	if err != nil {
		return nil, err
	}
	oldKey := refreshTokenPrefix + hashToken(refreshToken)
	newKey := refreshTokenPrefix + hashToken(newToken)
	result, err := rotateRefreshTokenScript.Run(
		ctx,
		s.rdb,
		[]string{oldKey, newKey},
		newToken,
		s.refreshTokenExpiry.Milliseconds(),
		refreshRotationGrace.Milliseconds(),
	).Slice()
	if err != nil || len(result) != 2 {
		return nil, apperr.ErrInvalidRefreshToken
	}
	status, ok := result[0].(int64)
	if !ok || status == 0 {
		return nil, apperr.ErrInvalidRefreshToken
	}
	value, ok := result[1].(string)
	if !ok || value == "" {
		return nil, apperr.ErrInvalidRefreshToken
	}

	userIDStr := value
	if status == 2 {
		newToken = value
		stored, getErr := s.rdb.Get(ctx, refreshTokenPrefix+hashToken(newToken)).Result()
		if getErr != nil || !strings.HasPrefix(stored, "user:") {
			return nil, apperr.ErrInvalidRefreshToken
		}
		userIDStr = strings.TrimPrefix(stored, "user:")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, apperr.ErrCorruptTokenData
	}

	user, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, apperr.ErrUserNotFound
	}
	if user.Email == nil {
		return nil, fmt.Errorf("user email is nil")
	}

	accessToken, err := s.generateJWT(userID, *user.Email)
	if err != nil {
		return nil, err
	}
	return &TokenRefreshResult{AccessToken: accessToken, RefreshToken: newToken}, nil
}

// RevokeRefreshToken removes a refresh token from Redis (used on logout).
func (s *AuthService) RevokeRefreshToken(ctx context.Context, refreshToken string) {
	key := refreshTokenPrefix + hashToken(refreshToken)
	keys := []string{key}
	if current, err := s.rdb.Get(ctx, key).Result(); err == nil && strings.HasPrefix(current, "rotated:") {
		replacement := strings.TrimPrefix(current, "rotated:")
		keys = append(keys, refreshTokenPrefix+hashToken(replacement))
	}
	if err := s.rdb.Del(ctx, keys...).Err(); err != nil {
		slog.Warn("Failed to revoke refresh token", "error", err)
	}
}
