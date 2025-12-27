package service

import (
	"context"
	"fmt"

	"cleaners-ai/internal/domain/entity"
	"cleaners-ai/internal/infrastructure/persistence"
	"cleaners-ai/pkg/auth"
)

type AuthService struct {
	userRepo     *persistence.UserRepository
	jwtManager   *auth.JWTManager
	googleOAuth  *auth.GoogleOAuthManager
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         *entity.User `json:"user"`
}

func NewAuthService(
	userRepo *persistence.UserRepository,
	jwtManager *auth.JWTManager,
	googleOAuth *auth.GoogleOAuthManager,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		jwtManager:  jwtManager,
		googleOAuth: googleOAuth,
	}
}

func (s *AuthService) GetGoogleAuthURL(state string) string {
	return s.googleOAuth.GetAuthURL(state)
}

func (s *AuthService) HandleGoogleCallback(ctx context.Context, code string) (*LoginResponse, error) {
	// Exchange code for token
	token, err := s.googleOAuth.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from Google
	googleUser, err := s.googleOAuth.GetUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Find or create user
	user, err := s.userRepo.GetByGoogleID(googleUser.ID)
	if err != nil {
		// User doesn't exist, create new one
		user = entity.NewUser(googleUser.Email, googleUser.Name, googleUser.Picture, googleUser.ID)
		if err := s.userRepo.Create(user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Generate JWT tokens
	accessToken, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (s *AuthService) RefreshToken(refreshTokenString string) (*LoginResponse, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get user
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Generate new tokens
	accessToken, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         user,
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*entity.User, error) {
	claims, err := s.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
