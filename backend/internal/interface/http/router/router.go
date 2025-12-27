package router

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"

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

	// -----------------------------
	// Routes
	// -----------------------------

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
		_, _ = w.Write([]byte("OK"))
	})

	// Healthz endpoint - returns 503 if DB is not connected
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unhealthy","db":"disconnected"}`))
			return
		}
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unhealthy","db":"error"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","db":"connected"}`))
	})

	// -----------------------------
	// Frontend static + SPA fallback
	// -----------------------------
	// Cloud Run 컨테이너 경로 기준으로 프론트 빌드 산출물을 서빙합니다.
	// 로그에서 Static directory: /app/frontend/build 로 찍혔으니 그 경로를 사용합니다.
	// (로컬 개발환경에서도 돌아가게 하려면 상대경로/환경변수로 바꿀 수 있습니다.)
	staticDir := "/app/frontend/build"
	mux.Handle("/", spaFileServer(staticDir))

	// CORS middleware wrapper (모든 응답에 CORS 헤더 추가)
	return enableCORS(mux)
}

// spaFileServer: React SPA 정적 서빙 + fallback(index.html)
// - 정적 파일이 있으면 그대로 제공
// - 없으면 index.html로 fallback (SPA 라우팅용)
// NOTE: API/Auth 라우트는 mux에서 먼저 처리되므로 여기서 별도 처리 불필요
func spaFileServer(staticDir string) http.Handler {
	fileServer := http.FileServer(http.Dir(staticDir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// 정적 파일 실제 경로 계산
		cleanPath := filepath.Clean(path)
		fullPath := filepath.Join(staticDir, cleanPath)

		// 디렉터리면 index.html
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}

		// 파일이 있으면 그대로 서빙
		if _, err := os.Stat(fullPath); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// 파일이 없으면 SPA fallback
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}

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
