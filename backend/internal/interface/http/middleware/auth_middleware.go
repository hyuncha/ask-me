package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/yourusername/cleaners-ai/pkg/auth"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const UserEmailKey contextKey = "user_email"

type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// RequireAuth is middleware that requires authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := m.extractToken(r)
		if token == "" {
			http.Error(w, `{"code":"UNAUTHORIZED","message":"No token provided"}`, http.StatusUnauthorized)
			return
		}

		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			http.Error(w, `{"code":"INVALID_TOKEN","message":"Invalid token"}`, http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth is middleware that optionally authenticates
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := m.extractToken(r)
		if token != "" {
			claims, err := m.jwtManager.ValidateToken(token)
			if err == nil {
				// Add user info to context
				ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
				ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (m *AuthMiddleware) extractToken(r *http.Request) string {
	// Try cookie first
	cookie, err := r.Cookie("access_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Try Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	return ""
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// GetUserEmailFromContext extracts user email from request context
func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}
