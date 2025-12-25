package service

import (
        "context"
        "fmt"

        "github.com/google/uuid"
        "github.com/yourusername/cleaners-ai/internal/domain/entity"
        "github.com/yourusername/cleaners-ai/internal/infrastructure/persistence"
        "github.com/yourusername/cleaners-ai/pkg/llm"
)

type ChatService struct {
        llmClient      *llm.OpenAIClient
        convRepo       *persistence.ConversationRepository
        messageRepo    *persistence.MessageRepository
        ragService     *RAGService
        anonymousUser  uuid.UUID
}

func NewChatService(
        llmClient *llm.OpenAIClient,
        convRepo *persistence.ConversationRepository,
        messageRepo *persistence.MessageRepository,
        ragService *RAGService,
) *ChatService {
        // For now, use a default anonymous user ID
        // In production, this should come from authentication
        anonymousUserID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

        return &ChatService{
                llmClient:     llmClient,
                convRepo:      convRepo,
                messageRepo:   messageRepo,
                ragService:    ragService,
                anonymousUser: anonymousUserID,
        }
}

func (s *ChatService) ProcessMessage(userMessage string, conversationID *uuid.UUID) (string, uuid.UUID, error) {
        if userMessage == "" {
                return "", uuid.Nil, fmt.Errorf("message cannot be empty")
        }

        // Check if DB is available
        dbAvailable := s.convRepo != nil && s.messageRepo != nil

        // Create or get conversation
        var convID uuid.UUID
        if conversationID == nil {
                if dbAvailable {
                        // Create new conversation in DB
                        conv := entity.NewConversation(s.anonymousUser, "New Chat", "ko")
                        if err := s.convRepo.Create(conv); err != nil {
                                return "", uuid.Nil, fmt.Errorf("failed to create conversation: %w", err)
                        }
                        convID = conv.ID
                } else {
                        // Generate a temporary conversation ID when DB is not available
                        convID = uuid.New()
                }
        } else {
                convID = *conversationID
        }

        // Save user message (only if DB available)
        if dbAvailable {
                userMsg := entity.NewMessage(convID, entity.RoleUser, userMessage)
                if err := s.messageRepo.Create(userMsg); err != nil {
                        return "", uuid.Nil, fmt.Errorf("failed to save user message: %w", err)
                }
        }

        // Get relevant context from RAG (if available)
        ctx := context.Background()
        var enhancedMessage string
        if s.ragService != nil {
                relevantContext, err := s.ragService.GetRelevantContext(ctx, userMessage)
                if err == nil && relevantContext != "" {
                        enhancedMessage = fmt.Sprintf("%s\n\n사용자 질문: %s", relevantContext, userMessage)
                } else {
                        enhancedMessage = userMessage
                }
        } else {
                enhancedMessage = userMessage
        }

        // Call OpenAI API with enhanced message
        response, err := s.llmClient.SendMessage(enhancedMessage)
        if err != nil {
                return "", uuid.Nil, fmt.Errorf("failed to get AI response: %w", err)
        }

        // Save assistant message (only if DB available)
        if dbAvailable {
                assistantMsg := entity.NewMessage(convID, entity.RoleAssistant, response)
                if err := s.messageRepo.Create(assistantMsg); err != nil {
                        return "", uuid.Nil, fmt.Errorf("failed to save assistant message: %w", err)
                }
        }

        return response, convID, nil
}
