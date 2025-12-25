package persistence

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/cleaners-ai/internal/domain/entity"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *entity.User) error {
	query := `
		INSERT INTO users (id, email, name, google_id, profile_image_url, subscription_tier, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(
		query,
		user.ID,
		user.Email,
		user.Name,
		user.GoogleID,
		user.ProfileImageURL,
		user.SubscriptionTier,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, email, name, google_id, profile_image_url, subscription_tier,
		       subscription_expires_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var user entity.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.GoogleID,
		&user.ProfileImageURL,
		&user.SubscriptionTier,
		&user.SubscriptionExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*entity.User, error) {
	query := `
		SELECT id, email, name, google_id, profile_image_url, subscription_tier,
		       subscription_expires_at, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var user entity.User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.GoogleID,
		&user.ProfileImageURL,
		&user.SubscriptionTier,
		&user.SubscriptionExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByGoogleID(googleID string) (*entity.User, error) {
	query := `
		SELECT id, email, name, google_id, profile_image_url, subscription_tier,
		       subscription_expires_at, created_at, updated_at
		FROM users
		WHERE google_id = $1
	`
	var user entity.User
	err := r.db.QueryRow(query, googleID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.GoogleID,
		&user.ProfileImageURL,
		&user.SubscriptionTier,
		&user.SubscriptionExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) Update(user *entity.User) error {
	query := `
		UPDATE users
		SET name = $1, profile_image_url = $2, subscription_tier = $3,
		    subscription_expires_at = $4, updated_at = $5
		WHERE id = $6
	`
	_, err := r.db.Exec(
		query,
		user.Name,
		user.ProfileImageURL,
		user.SubscriptionTier,
		user.SubscriptionExpiresAt,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}
