package service

import (
        "context"
        "fmt"

        "github.com/google/uuid"
        "github.com/yourusername/cleaners-ai/internal/domain/entity"
        "github.com/yourusername/cleaners-ai/internal/infrastructure/persistence"
        "github.com/yourusername/cleaners-ai/pkg/llm"
        "github.com/yourusername/cleaners-ai/pkg/vector"
)

type RAGService struct {
        knowledgeRepo   *persistence.KnowledgeRepository
        embeddingClient *llm.EmbeddingClient
        pineconeClient  *vector.PineconeClient
        namespace       string
}

func NewRAGService(
        knowledgeRepo *persistence.KnowledgeRepository,
        embeddingClient *llm.EmbeddingClient,
        pineconeClient *vector.PineconeClient,
        namespace string,
) *RAGService {
        return &RAGService{
                knowledgeRepo:   knowledgeRepo,
                embeddingClient: embeddingClient,
                pineconeClient:  pineconeClient,
                namespace:       namespace,
        }
}

// IndexKnowledge creates embeddings and stores in Pinecone
func (s *RAGService) IndexKnowledge(ctx context.Context, knowledge *entity.Knowledge) error {
        // If Pinecone client is not configured, just save to database
        if s.pineconeClient == nil {
                if err := s.knowledgeRepo.Create(ctx, knowledge); err != nil {
                        return fmt.Errorf("failed to save knowledge to database: %w", err)
                }
                return nil
        }

        // Create embedding for the content
        embedding, err := s.embeddingClient.CreateEmbedding(knowledge.Content)
        if err != nil {
                return fmt.Errorf("failed to create embedding: %w", err)
        }

        // Prepare vector with metadata
        vectorID := knowledge.ID.String()
        metadata := map[string]interface{}{
                "title":      knowledge.Title,
                "category":   string(knowledge.Category),
                "difficulty": string(knowledge.Difficulty),
                "language":   knowledge.Language,
                "tags":       knowledge.Tags,
        }

        vectors := []vector.Vector{
                {
                        ID:       vectorID,
                        Values:   embedding,
                        Metadata: metadata,
                },
        }

        // Upsert to Pinecone first
        if err := s.pineconeClient.Upsert(vectors, s.namespace); err != nil {
                return fmt.Errorf("failed to upsert to Pinecone: %w", err)
        }

        // Save to database after successful Pinecone upsert
        if err := s.knowledgeRepo.Create(ctx, knowledge); err != nil {
                return fmt.Errorf("failed to save knowledge to database: %w", err)
        }

        return nil
}

// SearchKnowledge finds relevant knowledge based on query
func (s *RAGService) SearchKnowledge(ctx context.Context, query string, topK int) ([]*entity.Knowledge, error) {
        // If Pinecone client is not configured, return empty results
        if s.pineconeClient == nil {
                return []*entity.Knowledge{}, nil
        }

        // Create embedding for the query
        queryEmbedding, err := s.embeddingClient.CreateEmbedding(query)
        if err != nil {
                return nil, fmt.Errorf("failed to create query embedding: %w", err)
        }

        // Query Pinecone
        filter := map[string]interface{}{
                "status": "active",
        }

        queryResp, err := s.pineconeClient.Query(queryEmbedding, topK, s.namespace, filter)
        if err != nil {
                return nil, fmt.Errorf("failed to query Pinecone: %w", err)
        }

        // Fetch full knowledge items from database
        var knowledgeItems []*entity.Knowledge
        for _, match := range queryResp.Matches {
                id, err := uuid.Parse(match.ID)
                if err != nil {
                        continue
                }

                knowledge, err := s.knowledgeRepo.GetByID(ctx, id)
                if err != nil {
                        continue
                }

                knowledgeItems = append(knowledgeItems, knowledge)
        }

        return knowledgeItems, nil
}

// GetRelevantContext retrieves relevant knowledge and formats as context
func (s *RAGService) GetRelevantContext(ctx context.Context, query string) (string, error) {
        knowledgeItems, err := s.SearchKnowledge(ctx, query, 3)
        if err != nil {
                return "", err
        }

        if len(knowledgeItems) == 0 {
                return "", nil
        }

        // Format knowledge items as context
        context := "관련 세탁 지식:\n\n"
        for i, item := range knowledgeItems {
                context += fmt.Sprintf("%d. %s\n%s\n\n", i+1, item.Title, item.Content)
        }

        return context, nil
}

// GetAllKnowledge returns all knowledge items from the database
func (s *RAGService) GetAllKnowledge(ctx context.Context) ([]*entity.Knowledge, error) {
        if s.knowledgeRepo == nil {
                return []*entity.Knowledge{}, nil
        }
        return s.knowledgeRepo.GetAll(ctx)
}

// DeleteKnowledge removes knowledge from both Pinecone and database
func (s *RAGService) DeleteKnowledge(ctx context.Context, id uuid.UUID) error {
        // Delete from Pinecone if configured
        if s.pineconeClient != nil {
                vectorID := id.String()
                if err := s.pineconeClient.Delete([]string{vectorID}, s.namespace); err != nil {
                        return fmt.Errorf("failed to delete from Pinecone: %w", err)
                }
        }

        // Delete from database
        if err := s.knowledgeRepo.Delete(ctx, id); err != nil {
                return fmt.Errorf("failed to delete from database: %w", err)
        }

        return nil
}
