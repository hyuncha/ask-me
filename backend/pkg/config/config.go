package config

import (
        "fmt"
        "os"
        "strconv"

        "github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
        Environment string
        Server      ServerConfig
        Database    DatabaseConfig
        Redis       RedisConfig
        Pinecone    PineconeConfig
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

        cfg := &Config{
                Environment: getEnv("ENVIRONMENT", "development"),
                Server: ServerConfig{
                        Port: getEnvAsInt("SERVER_PORT", 8080),
                        Host: getEnv("SERVER_HOST", "localhost"),
                },
                Database: DatabaseConfig{
                        Host:     getEnvWithFallback("DB_HOST", "PGHOST", "localhost"),
                        Port:     getEnvAsIntWithFallback("DB_PORT", "PGPORT", 5432),
                        User:     getEnvWithFallback("DB_USER", "PGUSER", "postgres"),
                        Password: getEnvWithFallback("DB_PASSWORD", "PGPASSWORD", ""),
                        DBName:   getEnvWithFallback("DB_NAME", "PGDATABASE", "cleaners_ai"),
                        SSLMode:  getEnv("DB_SSL_MODE", "require"),
                },
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
