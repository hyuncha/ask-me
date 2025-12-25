package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cleaners-ai/internal/application/service"
	"github.com/yourusername/cleaners-ai/internal/infrastructure/persistence"
)

type ChatHandler struct {
	chatService *service.ChatService
	convRepo    *persistence.ConversationRepository
	messageRepo *persistence.MessageRepository
}

func NewChatHandler(
	chatService *service.ChatService,
	convRepo *persistence.ConversationRepository,
	messageRepo *persistence.MessageRepository,
) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		convRepo:    convRepo,
		messageRepo: messageRepo,
	}
}

type SendMessageRequest struct {
	Message        string  `json:"message"`
	ConversationID *string `json:"conversation_id,omitempty"`
}

type SendMessageResponse struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id"`
	Timestamp      string `json:"timestamp"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SendMessage handles chat message requests
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Message == "" {
		h.sendError(w, http.StatusBadRequest, "EMPTY_MESSAGE", "Message cannot be empty")
		return
	}

	// Parse conversation ID if provided
	var convID *uuid.UUID
	if req.ConversationID != nil && *req.ConversationID != "" {
		parsed, err := uuid.Parse(*req.ConversationID)
		if err != nil {
			h.sendError(w, http.StatusBadRequest, "INVALID_CONVERSATION_ID", "Invalid conversation ID")
			return
		}
		convID = &parsed
	}

	// Call chat service to get AI response
	aiResponse, conversationID, err := h.chatService.ProcessMessage(req.Message, convID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "AI_ERROR", "Failed to process message: "+err.Error())
		return
	}

	response := SendMessageResponse{
		Message:        aiResponse,
		ConversationID: conversationID.String(),
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	h.sendJSON(w, http.StatusOK, response)
}

// GetConversationHistory gets messages for a conversation
func (h *ChatHandler) GetConversationHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Extract conversation ID from path
	conversationID := strings.TrimPrefix(r.URL.Path, "/api/chat/history/")
	if conversationID == "" {
		h.sendError(w, http.StatusBadRequest, "MISSING_CONVERSATION_ID", "Conversation ID required")
		return
	}

	convID, err := uuid.Parse(conversationID)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "INVALID_CONVERSATION_ID", "Invalid conversation ID")
		return
	}

	// Get messages
	messages, err := h.messageRepo.GetByConversationID(convID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch messages")
		return
	}

	h.sendJSON(w, http.StatusOK, messages)
}

// GetConversations gets all conversations for anonymous user
func (h *ChatHandler) GetConversations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// For now, use anonymous user ID
	anonymousUserID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	conversations, err := h.convRepo.GetByUserID(anonymousUserID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch conversations")
		return
	}

	h.sendJSON(w, http.StatusOK, conversations)
}

func (h *ChatHandler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *ChatHandler) sendError(w http.ResponseWriter, status int, code, message string) {
	h.sendJSON(w, status, ErrorResponse{
		Code:    code,
		Message: message,
	})
}
