package persistence

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"cleaners-ai/internal/domain/entity"
)

type ConversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) Create(conv *entity.Conversation) error {
	query := `
		INSERT INTO conversations (id, user_id, title, language, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(
		query,
		conv.ID,
		conv.UserID,
		conv.Title,
		conv.Language,
		conv.CreatedAt,
		conv.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}
	return nil
}

func (r *ConversationRepository) GetByID(id uuid.UUID) (*entity.Conversation, error) {
	query := `
		SELECT id, user_id, title, language, created_at, updated_at
		FROM conversations
		WHERE id = $1
	`
	var conv entity.Conversation
	err := r.db.QueryRow(query, id).Scan(
		&conv.ID,
		&conv.UserID,
		&conv.Title,
		&conv.Language,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("conversation not found")
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	return &conv, nil
}

func (r *ConversationRepository) GetByUserID(userID uuid.UUID) ([]*entity.Conversation, error) {
	query := `
		SELECT id, user_id, title, language, created_at, updated_at
		FROM conversations
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}
	defer rows.Close()

	var conversations []*entity.Conversation
	for rows.Next() {
		var conv entity.Conversation
		err := rows.Scan(
			&conv.ID,
			&conv.UserID,
			&conv.Title,
			&conv.Language,
			&conv.CreatedAt,
			&conv.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, &conv)
	}
	return conversations, nil
}

func (r *ConversationRepository) Update(conv *entity.Conversation) error {
	query := `
		UPDATE conversations
		SET title = $1, updated_at = $2
		WHERE id = $3
	`
	conv.UpdatedAt = time.Now()
	_, err := r.db.Exec(query, conv.Title, conv.UpdatedAt, conv.ID)
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}
	return nil
}

func (r *ConversationRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM conversations WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}
	return nil
}
