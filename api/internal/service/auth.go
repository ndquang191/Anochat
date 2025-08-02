package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/model"
	"golang.org/x/oauth2"
)

// AuthService handles authentication operations
type AuthService struct {
	oauthConfig *oauth2.Config
	jwtSecret   []byte
	userService *UserService
}

// GoogleUserInfo represents user data from Google OAuth
type GoogleUserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	Verified bool   `json:"email_verified"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// NewAuthService creates a new auth service
func NewAuthService(userService *UserService, oauthConfig *oauth2.Config, jwtSecret string) *AuthService {
	return &AuthService{
		oauthConfig: oauthConfig,
		jwtSecret:   []byte(jwtSecret),
		userService: userService,
	}
}

// ProcessOAuthCallback processes the OAuth callback and returns user with JWT
// This is the main function according to Backend.md specification
func (s *AuthService) ProcessOAuthCallback(ctx context.Context, code string) (*model.User, string, error) {
	// 1. Exchange code for token
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// 2. Get user info from Google
	googleUser, err := s.getUserInfoFromToken(ctx, token)
	if err != nil {
		return nil, "", err
	}

	// Debug: Log user info
	fmt.Printf("Google User Info: ID=%s, Email=%s, Name=%s, Verified=%v\n",
		googleUser.ID, googleUser.Email, googleUser.Name, googleUser.Verified)

	// 3. Verify email is verified (optional for development)
	if !googleUser.Verified {
		// For development, we can skip this check
		// In production, you might want to enforce email verification
		// return nil, "", fmt.Errorf("email not verified")
	}

	// 4. Get or create user in database
	user, err := s.userService.GetOrCreateUser(ctx, googleUser.Email, googleUser.Name, googleUser.Picture)
	if err != nil {
		return nil, "", err
	}

	// 5. Generate JWT token
	if user.Email == nil {
		return nil, "", fmt.Errorf("user email is nil")
	}
	jwtToken, err := s.generateJWT(user.ID, *user.Email)
	if err != nil {
		return nil, "", err
	}

	return user, jwtToken, nil
}

// GetUserFromToken gets user from JWT token and checks for active room
func (s *AuthService) GetUserFromToken(ctx context.Context, tokenString string) (*model.User, *model.Room, []model.Message, error) {
	// 1. Validate JWT and get claims
	claims, err := s.ValidateJWT(tokenString)
	if err != nil {
		return nil, nil, nil, err
	}

	// 2. Get user from database
	user, err := s.userService.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, nil, nil, err
	}

	// 3. Check if user has active room
	room, err := s.userService.GetActiveRoom(ctx, claims.UserID)
	if err != nil {
		// No active room found, return user only
		return user, nil, nil, nil
	}

	// 4. Get room messages if room exists
	messages, err := s.userService.GetRoomMessages(ctx, room.ID)
	if err != nil {
		return user, room, nil, err
	}

	return user, room, messages, nil
}

// getUserInfoFromToken gets user information from Google using access token
func (s *AuthService) getUserInfoFromToken(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := s.oauthConfig.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// generateJWT generates a JWT token for a user
func (s *AuthService) generateJWT(userID uuid.UUID, email string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour) // 24 hours

	claims := JWTClaims{
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

// ValidateJWT validates a JWT token and returns claims
func (s *AuthService) ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid JWT token")
}
