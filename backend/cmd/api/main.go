package main

import (
        "context"
        "database/sql"
        "fmt"
        "log"
        "net/http"
        "os"
        "os/signal"
        "syscall"
        "time"

        "cleaners-ai/internal/interface/http/router"
        "cleaners-ai/pkg/auth"
        "cleaners-ai/pkg/config"
        "cleaners-ai/pkg/database"
        "cleaners-ai/pkg/llm"
        "cleaners-ai/pkg/logger"
        "cleaners-ai/pkg/vector"
)

func main() {
        // Load configuration
        cfg, err := config.Load()
        if err != nil {
                log.Fatalf("Failed to load configuration: %v", err)
        }

        // Initialize logger
        log := logger.New(cfg.Environment)
        defer log.Sync()

        log.Info("Starting Cleaners AI API Server",
                "environment", cfg.Environment,
                "port", cfg.Server.Port,
                "llm_provider", cfg.LLM.Provider,
                "llm_model", cfg.LLM.Model,
        )

        // Initialize database connection (optional - server starts even if DB unavailable)
        var db *database.PostgresDB
        db, err = database.NewPostgresDB(cfg.Database)
        if err != nil {
                log.Warn("Failed to connect to database - server will start without DB",
                        "error", err,
                        "host", cfg.Database.Host,
                        "database", cfg.Database.DBName,
                )
        } else {
                defer db.Close()
                log.Info("Database connection established",
                        "host", cfg.Database.Host,
                        "database", cfg.Database.DBName,
                )

                // Run database migrations
                if err := db.RunMigrations(); err != nil {
                        log.Warn("Failed to run database migrations", "error", err)
                } else {
                        log.Info("Database migrations completed successfully")
                }
        }

        // Initialize LLM client
        llmClient := llm.NewOpenAIClient(cfg.LLM.APIKey, cfg.LLM.Model)
        log.Info("OpenAI client initialized")

        // Initialize embedding client
        embeddingClient := llm.NewEmbeddingClient(cfg.LLM.APIKey)
        log.Info("Embedding client initialized")

        // Initialize Pinecone client (only if credentials are provided)
        var pineconeClient *vector.PineconeClient
        if cfg.Pinecone.APIKey != "" && cfg.Pinecone.Environment != "" && cfg.Pinecone.IndexName != "" {
                pineconeClient = vector.NewPineconeClient(
                        cfg.Pinecone.APIKey,
                        cfg.Pinecone.Environment,
                        cfg.Pinecone.IndexName,
                )
                log.Info("Pinecone client initialized",
                        "environment", cfg.Pinecone.Environment,
                        "index", cfg.Pinecone.IndexName,
                )
        } else {
                log.Warn("Pinecone credentials not configured - RAG features will be limited")
        }

        // Initialize Pinecone Assistant client (for n8n-style RAG)
        var pineconeAssistant *vector.PineconeAssistantClient
        if cfg.Pinecone.APIKey != "" && cfg.Pinecone.AssistantID != "" {
                pineconeAssistant = vector.NewPineconeAssistantClient(
                        cfg.Pinecone.APIKey,
                        cfg.Pinecone.AssistantID,
                )
                log.Info("Pinecone Assistant client initialized",
                        "assistant_id", cfg.Pinecone.AssistantID,
                )
        }

        // Initialize OpenRouter client (if configured)
        var openRouterClient *llm.OpenRouterClient
        if cfg.OpenRouter.APIKey != "" {
                openRouterClient = llm.NewOpenRouterClient(
                        cfg.OpenRouter.APIKey,
                        cfg.OpenRouter.Model,
                )
                log.Info("OpenRouter client initialized",
                        "model", cfg.OpenRouter.Model,
                )
        }

        // Initialize JWT manager
        jwtManager := auth.NewJWTManager(
                cfg.Auth.JWTSecret,
                cfg.Auth.TokenExpiry,
                cfg.Auth.RefreshTokenExpiry,
        )
        log.Info("JWT manager initialized")

        // Initialize Google OAuth
        googleOAuth := auth.NewGoogleOAuthManager(auth.GoogleOAuthConfig{
                ClientID:     cfg.Auth.GoogleClientID,
                ClientSecret: cfg.Auth.GoogleClientSecret,
                RedirectURL:  cfg.Auth.GoogleRedirectURL,
        })
        log.Info("Google OAuth initialized")

        // Initialize router (pass nil if db is not available)
        var sqlDB *sql.DB
        if db != nil {
                sqlDB = db.DB
        }
        r := router.NewRouter(llmClient, sqlDB, jwtManager, googleOAuth, embeddingClient, pineconeClient, openRouterClient, pineconeAssistant)

        // Create HTTP server
        srv := &http.Server{
                Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
                Handler:      r,
                ReadTimeout:  15 * time.Second,
                WriteTimeout: 15 * time.Second,
                IdleTimeout:  60 * time.Second,
        }

        // Start server in a goroutine
        go func() {
                log.Info("Server is starting", "address", srv.Addr)
                if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
                        log.Fatal("Server failed to start", "error", err)
                }
        }()

        // Graceful shutdown
        quit := make(chan os.Signal, 1)
        signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
        <-quit

        log.Info("Server is shutting down...")

        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        if err := srv.Shutdown(ctx); err != nil {
                log.Fatal("Server forced to shutdown", "error", err)
        }

        log.Info("Server exited")
}
