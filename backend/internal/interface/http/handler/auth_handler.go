package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"

	"cleaners-ai/internal/application/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// GoogleLogin redirects to Google OAuth
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Generate state token for CSRF protection
	state := generateStateToken()

	// Store state in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	// Get Google OAuth URL
	authURL := h.authService.GetGoogleAuthURL(state)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GoogleCallback handles OAuth callback
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Verify state token
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "MISSING_STATE", "State cookie not found")
		return
	}

	state := r.URL.Query().Get("state")
	if state != stateCookie.Value {
		h.sendError(w, http.StatusBadRequest, "INVALID_STATE", "Invalid state token")
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		h.sendError(w, http.StatusBadRequest, "MISSING_CODE", "Authorization code not found")
		return
	}

	// Handle Google callback
	loginResp, err := h.authService.HandleGoogleCallback(r.Context(), code)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "AUTH_FAILED", "Authentication failed: "+err.Error())
		return
	}

	// Set tokens in cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    loginResp.AccessToken,
		MaxAge:   24 * 3600, // 24 hours
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    loginResp.RefreshToken,
		MaxAge:   7 * 24 * 3600, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	// Redirect to frontend (use FRONTEND_URL env var for Cloud Run deployment)
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000" // fallback for local development
	}
	http.Redirect(w, r, frontendURL, http.StatusTemporaryRedirect)
}

// RefreshToken refreshes the access token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	loginResp, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		h.sendError(w, http.StatusUnauthorized, "REFRESH_FAILED", "Failed to refresh token")
		return
	}

	h.sendJSON(w, http.StatusOK, loginResp)
}

// Logout clears auth cookies
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Clear cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})

	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// GetMe returns current user info
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Get token from cookie or Authorization header
	var token string
	cookie, err := r.Cookie("access_token")
	if err == nil {
		token = cookie.Value
	} else {
		// Try Authorization header
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	if token == "" {
		h.sendError(w, http.StatusUnauthorized, "UNAUTHORIZED", "No token provided")
		return
	}

	user, err := h.authService.ValidateToken(token)
	if err != nil {
		h.sendError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid token")
		return
	}

	h.sendJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *AuthHandler) sendError(w http.ResponseWriter, status int, code, message string) {
	h.sendJSON(w, status, ErrorResponse{
		Code:    code,
		Message: message,
	})
}

func generateStateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
