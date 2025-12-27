package config

import (
        "fmt"
        "net/url"
        "os"
        "strconv"
        "strings"

        "github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
        Environment string
        Server      ServerConfig
        Database    DatabaseConfig
        Redis       RedisConfig
        Pinecone    PineconeConfig
        OpenRouter  OpenRouterConfig
        Auth        AuthConfig
        LLM         LLMConfig
        Stripe      StripeConfig
}

type ServerConfig struct {
        Port int
        Host string
}

type DatabaseConfig struct {
        Host     string
        Port     int
        User     string
        Password string
        DBName   string
        SSLMode  string
}

type RedisConfig struct {
        Host     string
        Port     int
        Password string
        DB       int
}

type PineconeConfig struct {
        APIKey      string
        Environment string
        IndexName   string
        AssistantID string
}

type OpenRouterConfig struct {
        APIKey string
        Model  string
}

type AuthConfig struct {
        JWTSecret           string
        GoogleClientID      string
        GoogleClientSecret  string
        GoogleRedirectURL   string
        TokenExpiry         int // in hours
        RefreshTokenExpiry  int // in hours
}

type LLMConfig struct {
        Provider string // "openai", "anthropic", etc.
        APIKey   string
        Model    string
}

type StripeConfig struct {
        SecretKey      string
        PublishableKey string
        WebhookSecret  string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
        // Load .env file if exists (for local development)
        _ = godotenv.Load()

        // Parse DATABASE_URL if provided (highest priority)
        dbConfig, err := loadDatabaseConfig()
        if err != nil {
                return nil, err
        }

        cfg := &Config{
                Environment: getEnv("ENVIRONMENT", "development"),
                Server: ServerConfig{
                        Port: getEnvAsInt("SERVER_PORT", 8080),
                        Host: getEnv("SERVER_HOST", "0.0.0.0"), // 0.0.0.0 for Cloud Run compatibility
                },
                Database: dbConfig,
                Redis: RedisConfig{
                        Host:     getEnv("REDIS_HOST", "localhost"),
                        Port:     getEnvAsInt("REDIS_PORT", 6379),
                        Password: getEnv("REDIS_PASSWORD", ""),
                        DB:       getEnvAsInt("REDIS_DB", 0),
                },
                Pinecone: PineconeConfig{
                        APIKey:      getEnv("PINECONE_API_KEY", ""),
                        Environment: getEnv("PINECONE_ENV", ""),
                        IndexName:   getEnv("PINECONE_INDEX_NAME", "cleaners-knowledge"),
                        AssistantID: getEnv("PINECONE_ASSISTANT_ID", ""),
                },
                OpenRouter: OpenRouterConfig{
                        APIKey: getEnv("OPENROUTER_API_KEY", ""),
                        Model:  getEnv("OPENROUTER_MODEL", "openai/gpt-4.1"),
                },
                Auth: AuthConfig{
                        JWTSecret:          getEnv("JWT_SECRET", ""),
                        GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
                        GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
                        GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
                        TokenExpiry:        getEnvAsInt("TOKEN_EXPIRY_HOURS", 24),
                        RefreshTokenExpiry: getEnvAsInt("REFRESH_TOKEN_EXPIRY_HOURS", 168), // 7 days
                },
                LLM: LLMConfig{
                        Provider: getEnv("LLM_PROVIDER", "openai"),
                        APIKey:   getEnv("LLM_API_KEY", ""),
                        Model:    getEnv("LLM_MODEL", "gpt-4-turbo-preview"),
                },
                Stripe: StripeConfig{
                        SecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
                        PublishableKey: getEnv("STRIPE_PUBLISHABLE_KEY", ""),
                        WebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
                },
        }

        if err := cfg.Validate(); err != nil {
                return nil, err
        }

        return cfg, nil
}

// Validate validates required configuration values
func (c *Config) Validate() error {
        if c.LLM.APIKey == "" {
                return fmt.Errorf("LLM_API_KEY is required")
        }
        return nil
}

func getEnv(key, defaultValue string) string {
        if value := os.Getenv(key); value != "" {
                return value
        }
        return defaultValue
}

func getEnvWithFallback(primary, fallback, defaultValue string) string {
        if value := os.Getenv(primary); value != "" {
                return value
        }
        if value := os.Getenv(fallback); value != "" {
                return value
        }
        return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
        valueStr := os.Getenv(key)
        if value, err := strconv.Atoi(valueStr); err == nil {
                return value
        }
        return defaultValue
}

func getEnvAsIntWithFallback(primary, fallback string, defaultValue int) int {
        if valueStr := os.Getenv(primary); valueStr != "" {
                if value, err := strconv.Atoi(valueStr); err == nil {
                        return value
                }
        }
        if valueStr := os.Getenv(fallback); valueStr != "" {
                if value, err := strconv.Atoi(valueStr); err == nil {
                        return value
                }
        }
        return defaultValue
}

// loadDatabaseConfig loads database configuration with strict validation
// Priority: DATABASE_URL > PGHOST/DB_HOST env vars
// NO localhost fallback - missing config is a fatal error
func loadDatabaseConfig() (DatabaseConfig, error) {
        // Priority 1: DATABASE_URL (supports both TCP and Unix socket)
        if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
                return parseDatabaseURL(dbURL)
        }

        // Priority 2: Individual environment variables (PGHOST or DB_HOST)
        host := getEnvWithFallback("DB_HOST", "PGHOST", "")
        if host == "" {
                return DatabaseConfig{}, fmt.Errorf("DATABASE_URL or DB_HOST/PGHOST is required - localhost fallback is disabled")
        }

        // Validate: reject localhost explicitly
        if host == "localhost" || host == "127.0.0.1" {
                return DatabaseConfig{}, fmt.Errorf("localhost database connection is not allowed in production - use Cloud SQL socket or remote host")
        }

        user := getEnvWithFallback("DB_USER", "PGUSER", "")
        if user == "" {
                return DatabaseConfig{}, fmt.Errorf("DB_USER/PGUSER is required")
        }

        password := getEnvWithFallback("DB_PASSWORD", "PGPASSWORD", "")
        dbName := getEnvWithFallback("DB_NAME", "PGDATABASE", "")
        if dbName == "" {
                return DatabaseConfig{}, fmt.Errorf("DB_NAME/PGDATABASE is required")
        }

        // SSLMode: for Unix socket (Cloud SQL), disable is OK (socket is already secure)
        // For TCP connections, require SSL
        sslMode := getEnvWithFallback("DB_SSL_MODE", "PGSSLMODE", "require")
        if sslMode == "disable" && !strings.HasPrefix(host, "/") {
                sslMode = "require" // Force SSL for TCP connections
        }

        port := getEnvAsIntWithFallback("DB_PORT", "PGPORT", 5432)

        config := DatabaseConfig{
                Host:     host,
                Port:     port,
                User:     user,
                Password: password,
                DBName:   dbName,
                SSLMode:  sslMode,
        }

        // Log connection target (no secrets)
        fmt.Printf("[DB Config] host=%s port=%d db=%s sslmode=%s\n", host, port, dbName, sslMode)

        return config, nil
}

// parseDatabaseURL parses DATABASE_URL supporting both TCP and Unix socket formats
// TCP: postgres://user:pass@host:port/dbname?sslmode=require
// Unix: postgres://user:pass@/dbname?host=/cloudsql/project:region:instance
func parseDatabaseURL(dbURL string) (DatabaseConfig, error) {
        u, err := url.Parse(dbURL)
        if err != nil {
                return DatabaseConfig{}, fmt.Errorf("invalid DATABASE_URL: %w", err)
        }

        if u.Scheme != "postgres" && u.Scheme != "postgresql" {
                return DatabaseConfig{}, fmt.Errorf("DATABASE_URL must use postgres:// scheme")
        }

        user := u.User.Username()
        password, _ := u.User.Password()
        dbName := strings.TrimPrefix(u.Path, "/")

        // Parse query parameters
        query := u.Query()
        sslMode := query.Get("sslmode")

        var host string
        var port int

        // Check for Unix socket in query parameter (Cloud SQL style)
        if socketHost := query.Get("host"); socketHost != "" && strings.HasPrefix(socketHost, "/") {
                // Unix socket: host=/cloudsql/project:region:instance
                host = socketHost
                port = 5432 // Port is ignored for Unix sockets but keep for struct
                // Unix socket is already secure, sslmode=disable is OK
                if sslMode == "" {
                        sslMode = "disable"
                }
        } else if u.Host != "" {
                // TCP connection
                host = u.Hostname()
                portStr := u.Port()
                if portStr != "" {
                        port, _ = strconv.Atoi(portStr)
                } else {
                        port = 5432
                }
                // TCP connection requires SSL
                if sslMode == "" || sslMode == "disable" {
                        sslMode = "require"
                }
        } else {
                return DatabaseConfig{}, fmt.Errorf("DATABASE_URL must specify host or socket path")
        }

        // Validate: reject localhost
        if host == "localhost" || host == "127.0.0.1" {
                return DatabaseConfig{}, fmt.Errorf("localhost database connection is not allowed - use Cloud SQL socket or remote host")
        }

        if user == "" {
                return DatabaseConfig{}, fmt.Errorf("DATABASE_URL must include username")
        }
        if dbName == "" {
                return DatabaseConfig{}, fmt.Errorf("DATABASE_URL must include database name")
        }

        config := DatabaseConfig{
                Host:     host,
                Port:     port,
                User:     user,
                Password: password,
                DBName:   dbName,
                SSLMode:  sslMode,
        }

        // Log connection target (no secrets)
        if strings.HasPrefix(host, "/") {
                fmt.Printf("[DB Config] socket=%s db=%s sslmode=%s\n", host, dbName, sslMode)
        } else {
                fmt.Printf("[DB Config] host=%s port=%d db=%s sslmode=%s\n", host, port, dbName, sslMode)
        }

        return config, nil
}
