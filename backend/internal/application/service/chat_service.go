package service

import (
        "context"
        "fmt"
        "sync"

        "github.com/google/uuid"
        "cleaners-ai/internal/domain/entity"
        "cleaners-ai/internal/infrastructure/persistence"
        "cleaners-ai/pkg/llm"
        "cleaners-ai/pkg/vector"
)

// SessionMessage represents a message in the session memory
type SessionMessage struct {
        Role    string
        Content string
}

// SessionMemory manages in-memory conversation history per session
type SessionMemory struct {
        mu       sync.RWMutex
        sessions map[string][]SessionMessage
        maxSize  int
}

// NewSessionMemory creates a new session memory manager
func NewSessionMemory(maxSize int) *SessionMemory {
        if maxSize <= 0 {
                maxSize = 10 // Default to 10 messages per session
        }
        return &SessionMemory{
                sessions: make(map[string][]SessionMessage),
                maxSize:  maxSize,
        }
}

// AddMessage adds a message to a session
func (m *SessionMemory) AddMessage(sessionID, role, content string) {
        m.mu.Lock()
        defer m.mu.Unlock()

        if _, exists := m.sessions[sessionID]; !exists {
                m.sessions[sessionID] = make([]SessionMessage, 0, m.maxSize)
        }

        m.sessions[sessionID] = append(m.sessions[sessionID], SessionMessage{
                Role:    role,
                Content: content,
        })

        // Trim to max size (keep most recent messages)
        if len(m.sessions[sessionID]) > m.maxSize {
                m.sessions[sessionID] = m.sessions[sessionID][len(m.sessions[sessionID])-m.maxSize:]
        }
}

// GetHistory returns the conversation history for a session
func (m *SessionMemory) GetHistory(sessionID string) []SessionMessage {
        m.mu.RLock()
        defer m.mu.RUnlock()

        if history, exists := m.sessions[sessionID]; exists {
                // Return a copy to prevent race conditions
                result := make([]SessionMessage, len(history))
                copy(result, history)
                return result
        }
        return nil
}

// ClearSession clears the history for a session
func (m *SessionMemory) ClearSession(sessionID string) {
        m.mu.Lock()
        defer m.mu.Unlock()
        delete(m.sessions, sessionID)
}

// LaundryMessageResult contains the AI response and recommended shops
type LaundryMessageResult struct {
        Message          string
        RecommendedShops []PartnerShop
}

type ChatService struct {
        llmClient          *llm.OpenAIClient
        openRouterClient   *llm.OpenRouterClient
        pineconeAssistant  *vector.PineconeAssistantClient
        convRepo           *persistence.ConversationRepository
        messageRepo        *persistence.MessageRepository
        ragService         *RAGService
        partnerShopService *PartnerShopService
        sessionMemory      *SessionMemory
        anonymousUser      uuid.UUID
        useOpenRouter      bool
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
                llmClient:          llmClient,
                convRepo:           convRepo,
                messageRepo:        messageRepo,
                ragService:         ragService,
                partnerShopService: NewPartnerShopService(nil),
                sessionMemory:      NewSessionMemory(10),
                anonymousUser:      anonymousUserID,
                useOpenRouter:      false,
        }
}

// NewChatServiceWithOpenRouter creates a ChatService that uses OpenRouter and Pinecone Assistant
func NewChatServiceWithOpenRouter(
        openRouterClient *llm.OpenRouterClient,
        pineconeAssistant *vector.PineconeAssistantClient,
        convRepo *persistence.ConversationRepository,
        messageRepo *persistence.MessageRepository,
        ragService *RAGService,
) *ChatService {
        anonymousUserID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

        return &ChatService{
                openRouterClient:   openRouterClient,
                pineconeAssistant:  pineconeAssistant,
                convRepo:           convRepo,
                messageRepo:        messageRepo,
                ragService:         ragService,
                partnerShopService: NewPartnerShopService(pineconeAssistant),
                sessionMemory:      NewSessionMemory(10),
                anonymousUser:      anonymousUserID,
                useOpenRouter:      true,
        }
}

// SetOpenRouterClient sets the OpenRouter client
func (s *ChatService) SetOpenRouterClient(client *llm.OpenRouterClient) {
        s.openRouterClient = client
        s.useOpenRouter = client != nil
}

// SetPineconeAssistant sets the Pinecone Assistant client
func (s *ChatService) SetPineconeAssistant(client *vector.PineconeAssistantClient) {
        s.pineconeAssistant = client
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

// ProcessMessageWithSession handles messages with session-based memory (n8n workflow style)
func (s *ChatService) ProcessMessageWithSession(userMessage string, sessionID string) (string, error) {
        if userMessage == "" {
                return "", fmt.Errorf("message cannot be empty")
        }

        if sessionID == "" {
                sessionID = uuid.New().String()
        }

        // Add user message to session memory
        s.sessionMemory.AddMessage(sessionID, "user", userMessage)

        var response string
        var err error

        // Get context from Pinecone Assistant if available
        var pineconeContext string
        if s.pineconeAssistant != nil {
                pineconeContext, err = s.pineconeAssistant.Search(userMessage)
                if err != nil {
                        // Log error but continue without Pinecone context
                        pineconeContext = ""
                }
        }

        // Get conversation history
        history := s.sessionMemory.GetHistory(sessionID)

        if s.useOpenRouter && s.openRouterClient != nil {
                // Convert session history to OpenRouter messages
                openRouterHistory := make([]llm.OpenRouterMessage, 0, len(history)-1)
                for i := 0; i < len(history)-1; i++ { // Exclude the last message (current user message)
                        openRouterHistory = append(openRouterHistory, llm.OpenRouterMessage{
                                Role:    history[i].Role,
                                Content: history[i].Content,
                        })
                }

                // Use OpenRouter with history
                response, err = s.openRouterClient.SendMessageWithHistory(userMessage, pineconeContext, openRouterHistory)
        } else if s.llmClient != nil {
                // Fallback to OpenAI client
                if pineconeContext != "" {
                        response, err = s.llmClient.SendMessageWithContext(userMessage, pineconeContext)
                } else {
                        response, err = s.llmClient.SendMessage(userMessage)
                }
        } else {
                return "", fmt.Errorf("no LLM client configured")
        }

        if err != nil {
                return "", fmt.Errorf("failed to get AI response: %w", err)
        }

        // Add assistant response to session memory
        s.sessionMemory.AddMessage(sessionID, "assistant", response)

        return response, nil
}

// ClearSession clears the conversation history for a session
func (s *ChatService) ClearSession(sessionID string) {
        s.sessionMemory.ClearSession(sessionID)
}

// ProcessLaundryMessage handles laundry-related messages with partner shop recommendations
func (s *ChatService) ProcessLaundryMessage(userMessage string, sessionID string, location string) (*LaundryMessageResult, error) {
        if userMessage == "" {
                return nil, fmt.Errorf("message cannot be empty")
        }

        if sessionID == "" {
                sessionID = uuid.New().String()
        }

        // Add user message to session memory
        s.sessionMemory.AddMessage(sessionID, "user", userMessage)

        var response string
        var err error

        // Get context from Pinecone Assistant if available
        var pineconeContext string
        if s.pineconeAssistant != nil {
                pineconeContext, err = s.pineconeAssistant.Search(userMessage)
                if err != nil {
                        // Log error but continue without Pinecone context
                        pineconeContext = ""
                }
        }

        // Get conversation history
        history := s.sessionMemory.GetHistory(sessionID)

        if s.useOpenRouter && s.openRouterClient != nil {
                // Convert session history to OpenRouter messages
                openRouterHistory := make([]llm.OpenRouterMessage, 0, len(history)-1)
                for i := 0; i < len(history)-1; i++ { // Exclude the last message (current user message)
                        openRouterHistory = append(openRouterHistory, llm.OpenRouterMessage{
                                Role:    history[i].Role,
                                Content: history[i].Content,
                        })
                }

                // Use OpenRouter with history
                response, err = s.openRouterClient.SendMessageWithHistory(userMessage, pineconeContext, openRouterHistory)
        } else if s.llmClient != nil {
                // Fallback to OpenAI client
                if pineconeContext != "" {
                        response, err = s.llmClient.SendMessageWithContext(userMessage, pineconeContext)
                } else {
                        response, err = s.llmClient.SendMessage(userMessage)
                }
        } else {
                return nil, fmt.Errorf("no LLM client configured")
        }

        if err != nil {
                return nil, fmt.Errorf("failed to get AI response: %w", err)
        }

        // Add assistant response to session memory
        s.sessionMemory.AddMessage(sessionID, "assistant", response)

        // Check if we should recommend partner shops
        var recommendedShops []PartnerShop
        if s.partnerShopService != nil && s.partnerShopService.ShouldRecommendShop(userMessage, 0) {
                shops, shopErr := s.partnerShopService.GetPartnerShopsByLocation(location)
                if shopErr == nil {
                        recommendedShops = shops
                }
        }

        return &LaundryMessageResult{
                Message:          response,
                RecommendedShops: recommendedShops,
        }, nil
}
