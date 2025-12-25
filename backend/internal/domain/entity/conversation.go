package entity

import (
	"time"

	"github.com/google/uuid"
)

// MessageRole represents the role of message sender
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleSystem    MessageRole = "system"
)

// Conversation represents a chat conversation
type Conversation struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message represents a single message in a conversation
type Message struct {
	ID             uuid.UUID   `json:"id"`
	ConversationID uuid.UUID   `json:"conversation_id"`
	Role           MessageRole `json:"role"`
	Content        string      `json:"content"`
	CreatedAt      time.Time   `json:"created_at"`
}

// NewConversation creates a new conversation
func NewConversation(userID uuid.UUID, title, language string) *Conversation {
	now := time.Now()
	return &Conversation{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		Language:  language,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewMessage creates a new message
func NewMessage(conversationID uuid.UUID, role MessageRole, content string) *Message {
	return &Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		CreatedAt:      time.Now(),
	}
}
