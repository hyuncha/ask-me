package router

import (
        "database/sql"
        "io/fs"
        "net/http"
        "os"
        "path/filepath"
        "strings"

        "cleaners-ai/internal/application/service"
        "cleaners-ai/internal/infrastructure/persistence"
        "cleaners-ai/internal/interface/http/handler"
        "cleaners-ai/pkg/auth"
        "cleaners-ai/pkg/llm"
        "cleaners-ai/pkg/vector"
)

func NewRouter(
        llmClient *llm.OpenAIClient,
        db *sql.DB,
        jwtManager *auth.JWTManager,
        googleOAuth *auth.GoogleOAuthManager,
        embeddingClient *llm.EmbeddingClient,
        pineconeClient *vector.PineconeClient,
        openRouterClient *llm.OpenRouterClient,
        pineconeAssistant *vector.PineconeAssistantClient,
) http.Handler {
        mux := http.NewServeMux()

        // Initialize repositories (only if DB is available)
        var convRepo *persistence.ConversationRepository
        var messageRepo *persistence.MessageRepository
        var userRepo *persistence.UserRepository
        var knowledgeRepo *persistence.KnowledgeRepository

        if db != nil {
                convRepo = persistence.NewConversationRepository(db)
                messageRepo = persistence.NewMessageRepository(db)
                userRepo = persistence.NewUserRepository(db)
                knowledgeRepo = persistence.NewKnowledgeRepository(db)
        }

        // Initialize services
        ragService := service.NewRAGService(knowledgeRepo, embeddingClient, pineconeClient, "cleaners-ai")
        chatService := service.NewChatService(llmClient, convRepo, messageRepo, ragService)

        // Configure OpenRouter and Pinecone Assistant if available
        if openRouterClient != nil {
                chatService.SetOpenRouterClient(openRouterClient)
        }
        if pineconeAssistant != nil {
                chatService.SetPineconeAssistant(pineconeAssistant)
        }

        authService := service.NewAuthService(userRepo, jwtManager, googleOAuth)

        // Initialize handlers
        chatHandler := handler.NewChatHandler(chatService, convRepo, messageRepo)
        authHandler := handler.NewAuthHandler(authService)
        uploadHandler := handler.NewUploadHandler("./uploads")
        knowledgeHandler := handler.NewKnowledgeHandler(ragService)
        textExtractionHandler := handler.NewTextExtractionHandler()

        // CORS middleware wrapper
        corsHandler := enableCORS(mux)

        // Auth routes
        mux.HandleFunc("/auth/google", authHandler.GoogleLogin)
        mux.HandleFunc("/auth/google/callback", authHandler.GoogleCallback)
        mux.HandleFunc("/auth/refresh", authHandler.RefreshToken)
        mux.HandleFunc("/auth/logout", authHandler.Logout)
        mux.HandleFunc("/auth/me", authHandler.GetMe)

        // Chat routes
        mux.HandleFunc("/api/chat/message", chatHandler.SendMessage)
        mux.HandleFunc("/api/chat/conversations", chatHandler.GetConversations)
        mux.HandleFunc("/api/chat/history/", chatHandler.GetConversationHistory)

        // Upload routes
        mux.HandleFunc("/api/upload", uploadHandler.UploadFile)
        mux.HandleFunc("/uploads/", uploadHandler.ServeFile)

        // Knowledge routes
        mux.HandleFunc("/api/knowledge", func(w http.ResponseWriter, r *http.Request) {
                if r.Method == http.MethodGet {
                        knowledgeHandler.ListKnowledge(w, r)
                } else if r.Method == http.MethodPost {
                        knowledgeHandler.CreateKnowledge(w, r)
                } else {
                        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                }
        })
        mux.HandleFunc("/api/knowledge/search", knowledgeHandler.SearchKnowledge)
        mux.HandleFunc("/api/knowledge/", knowledgeHandler.DeleteKnowledge)

        // Text extraction route
        mux.HandleFunc("/api/extract-text", textExtractionHandler.ExtractText)

        // Health check with DB status
        mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
                w.Write([]byte("OK"))
        })

        // Healthz endpoint - returns 503 if DB is not connected
        mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
                if db == nil {
                        w.WriteHeader(http.StatusServiceUnavailable)
                        w.Write([]byte(`{"status":"unhealthy","db":"disconnected"}`))
                        return
                }
                if err := db.Ping(); err != nil {
                        w.WriteHeader(http.StatusServiceUnavailable)
                        w.Write([]byte(`{"status":"unhealthy","db":"error"}`))
                        return
                }
                w.WriteHeader(http.StatusOK)
                w.Write([]byte(`{"status":"healthy","db":"connected"}`))
        })

        // Serve static frontend files (for production)
        staticDir := os.Getenv("STATIC_DIR")
        if staticDir == "" {
                staticDir = "../frontend/build"
        }
        if _, err := os.Stat(staticDir); err == nil {
                mux.HandleFunc("/", createSPAHandler(staticDir))
        }

        return corsHandler
}

// createSPAHandler creates a handler for serving Single Page Application files
func createSPAHandler(staticDir string) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                // Don't serve SPA for API routes
                if strings.HasPrefix(r.URL.Path, "/api/") ||
                        strings.HasPrefix(r.URL.Path, "/auth/") ||
                        strings.HasPrefix(r.URL.Path, "/uploads/") ||
                        r.URL.Path == "/health" {
                        http.NotFound(w, r)
                        return
                }

                // Get the requested path
                path := filepath.Join(staticDir, r.URL.Path)

                // Check if file exists
                fi, err := os.Stat(path)
                if os.IsNotExist(err) || (err == nil && fi.IsDir()) {
                        // Serve index.html for SPA routing
                        http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
                        return
                }

                // Serve the actual file
                http.ServeFile(w, r, path)
        }
}

// Needed to prevent unused import error
var _ = fs.FS(nil)

// enableCORS adds CORS headers to all responses
func enableCORS(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Access-Control-Allow-Origin", "*")
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                w.Header().Set("Access-Control-Allow-Credentials", "true")

                // Handle preflight requests
                if r.Method == http.MethodOptions {
                        w.WriteHeader(http.StatusOK)
                        return
                }

                next.ServeHTTP(w, r)
        })
}
