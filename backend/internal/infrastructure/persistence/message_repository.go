package persistence

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"cleaners-ai/internal/domain/entity"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(msg *entity.Message) error {
	query := `
		INSERT INTO messages (id, conversation_id, role, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(
		query,
		msg.ID,
		msg.ConversationID,
		msg.Role,
		msg.Content,
		msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}
	return nil
}

func (r *MessageRepository) GetByConversationID(conversationID uuid.UUID) ([]*entity.Message, error) {
	query := `
		SELECT id, conversation_id, role, content, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []*entity.Message
	for rows.Next() {
		var msg entity.Message
		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.Role,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, &msg)
	}
	return messages, nil
}

func (r *MessageRepository) GetRecentMessages(conversationID uuid.UUID, limit int) ([]*entity.Message, error) {
	query := `
		SELECT id, conversation_id, role, content, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(query, conversationID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent messages: %w", err)
	}
	defer rows.Close()

	var messages []*entity.Message
	for rows.Next() {
		var msg entity.Message
		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.Role,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, &msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *MessageRepository) DeleteByConversationID(conversationID uuid.UUID) error {
	query := `DELETE FROM messages WHERE conversation_id = $1`
	_, err := r.db.Exec(query, conversationID)
	if err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}
	return nil
}
