package handler

import (
        "encoding/json"
        "net/http"
        "strings"

        "github.com/google/uuid"
        "github.com/yourusername/cleaners-ai/internal/application/service"
        "github.com/yourusername/cleaners-ai/internal/domain/entity"
)

type KnowledgeHandler struct {
        ragService *service.RAGService
}

func NewKnowledgeHandler(ragService *service.RAGService) *KnowledgeHandler {
        return &KnowledgeHandler{
                ragService: ragService,
        }
}

type CreateKnowledgeRequest struct {
        Title      string                      `json:"title"`
        Content    string                      `json:"content"`
        Category   entity.KnowledgeCategory    `json:"category"`
        Difficulty entity.KnowledgeDifficulty  `json:"difficulty"`
        Tags       []string                    `json:"tags"`
        Language   string                      `json:"language"`
}

type SearchKnowledgeRequest struct {
        Query string `json:"query"`
        TopK  int    `json:"top_k"`
}

// ListKnowledge handles GET /api/knowledge
func (h *KnowledgeHandler) ListKnowledge(w http.ResponseWriter, r *http.Request) {
        knowledgeItems, err := h.ragService.GetAllKnowledge(r.Context())
        if err != nil {
                http.Error(w, "Failed to list knowledge: "+err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
                "items": knowledgeItems,
                "count": len(knowledgeItems),
        })
}

// CreateKnowledge handles POST /api/knowledge
func (h *KnowledgeHandler) CreateKnowledge(w http.ResponseWriter, r *http.Request) {
        var req CreateKnowledgeRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                http.Error(w, "Invalid request body", http.StatusBadRequest)
                return
        }

        // Validate required fields
        if req.Title == "" || req.Content == "" {
                http.Error(w, "Title and content are required", http.StatusBadRequest)
                return
        }

        // Default to anonymous user for now (in production, get from auth context)
        createdBy := uuid.MustParse("00000000-0000-0000-0000-000000000000")

        // Default language to Korean if not specified
        if req.Language == "" {
                req.Language = "KR"
        }

        // Create knowledge entity
        knowledge := entity.NewKnowledge(
                req.Title,
                req.Content,
                req.Category,
                req.Difficulty,
                req.Tags,
                req.Language,
                createdBy,
        )

        // Index knowledge (save to DB and Pinecone)
        if err := h.ragService.IndexKnowledge(r.Context(), knowledge); err != nil {
                http.Error(w, "Failed to create knowledge: "+err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]interface{}{
                "id":      knowledge.ID,
                "message": "Knowledge created successfully",
        })
}

// SearchKnowledge handles POST /api/knowledge/search
func (h *KnowledgeHandler) SearchKnowledge(w http.ResponseWriter, r *http.Request) {
        var req SearchKnowledgeRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                http.Error(w, "Invalid request body", http.StatusBadRequest)
                return
        }

        if req.Query == "" {
                http.Error(w, "Query is required", http.StatusBadRequest)
                return
        }

        // Default topK to 5 if not specified
        if req.TopK <= 0 {
                req.TopK = 5
        }

        // Search knowledge
        knowledgeItems, err := h.ragService.SearchKnowledge(r.Context(), req.Query, req.TopK)
        if err != nil {
                http.Error(w, "Failed to search knowledge: "+err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
                "results": knowledgeItems,
                "count":   len(knowledgeItems),
        })
}

// DeleteKnowledge handles DELETE /api/knowledge/{id}
func (h *KnowledgeHandler) DeleteKnowledge(w http.ResponseWriter, r *http.Request) {
        // Check HTTP method
        if r.Method != http.MethodDelete {
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                return
        }

        // Extract ID from URL path: /api/knowledge/{id}
        path := strings.TrimPrefix(r.URL.Path, "/api/knowledge/")
        idStr := strings.TrimSuffix(path, "/")
        
        if idStr == "" {
                http.Error(w, "Knowledge ID is required", http.StatusBadRequest)
                return
        }

        id, err := uuid.Parse(idStr)
        if err != nil {
                http.Error(w, "Invalid knowledge ID: "+idStr, http.StatusBadRequest)
                return
        }

        // Delete knowledge
        if err := h.ragService.DeleteKnowledge(r.Context(), id); err != nil {
                http.Error(w, "Failed to delete knowledge: "+err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
                "message": "Knowledge deleted successfully",
        })
}
