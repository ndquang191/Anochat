package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"golang.org/x/oauth2"
)

type AuthService struct {
	oauthConfig *oauth2.Config
	jwtSecret   []byte
	userService *UserService
}

func NewAuthService(
	userService *UserService,
	oauthConfig *oauth2.Config,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		oauthConfig: oauthConfig,
		jwtSecret:   []byte(jwtSecret),
		userService: userService,
	}
}

func (s *AuthService) ProcessOAuthCallback(ctx context.Context, code string) (*identity.User, string, error) {
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("failed to exchange code for token: %w", err)
	}

	googleUser, err := s.getUserInfoFromToken(ctx, token)
	if err != nil {
		return nil, "", err
	}

	slog.Info("Google OAuth user info received", "email", googleUser.Email, "name", googleUser.Name)

	user, err := s.userService.GetOrCreateUser(ctx, googleUser.Email, googleUser.Name, googleUser.Picture)
	if err != nil {
		return nil, "", err
	}

	// Auto-calculate age from Google birthday (best-effort, silently ignored if unavailable)
	if age := s.getAgeFromGoogle(ctx, token); age != nil {
		s.userService.UpdateProfile(ctx, user.ID, nil, nil, age, nil)
	}

	if user.Email == nil {
		return nil, "", fmt.Errorf("user email is nil")
	}
	jwtToken, err := s.generateJWT(user.ID, *user.Email)
	if err != nil {
		return nil, "", err
	}

	return user, jwtToken, nil
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
	if err != nil || resp.StatusCode != http.StatusOK {
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
	expiresAt := now.Add(24 * time.Hour)

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
	return nil, fmt.Errorf("invalid JWT token")
}
