package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents user role in the system
type UserRole string

const (
	RoleConsumer UserRole = "consumer"
	RoleOwner    UserRole = "owner"
	RoleAdmin    UserRole = "admin"
)

// User represents a user entity
type User struct {
	ID                    uuid.UUID  `json:"id"`
	Email                 string     `json:"email"`
	Name                  string     `json:"name"`
	Picture               string     `json:"picture"`
	ProfileImageURL       string     `json:"profile_image_url"`
	Role                  UserRole   `json:"role"`
	GoogleID              string     `json:"google_id"`
	Language              string     `json:"language"` // "KR" or "EN"
	SubscriptionTier      string     `json:"subscription_tier"`
	SubscriptionExpiresAt *time.Time `json:"subscription_expires_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	LastLoginAt           *time.Time `json:"last_login_at,omitempty"`
}

// NewUser creates a new user
func NewUser(email, name, picture, googleID string) *User {
	now := time.Now()
	return &User{
		ID:               uuid.New(),
		Email:            email,
		Name:             name,
		Picture:          picture,
		ProfileImageURL:  picture,
		Role:             RoleConsumer, // Default role
		GoogleID:         googleID,
		Language:         "KR", // Default language
		SubscriptionTier: "free",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// IsAdmin checks if user is an admin
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsOwner checks if user is a laundry shop owner
func (u *User) IsOwner() bool {
	return u.Role == RoleOwner
}

// IsConsumer checks if user is a consumer
func (u *User) IsConsumer() bool {
	return u.Role == RoleConsumer
}
