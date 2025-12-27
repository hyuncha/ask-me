package database

import (
        "database/sql"
        "fmt"
        "time"

        _ "github.com/lib/pq"
        "cleaners-ai/pkg/config"
)

type PostgresDB struct {
        *sql.DB
}

func NewPostgresDB(cfg config.DatabaseConfig) (*PostgresDB, error) {
        dsn := fmt.Sprintf(
                "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
                cfg.Host,
                cfg.Port,
                cfg.User,
                cfg.Password,
                cfg.DBName,
                cfg.SSLMode,
        )

        db, err := sql.Open("postgres", dsn)
        if err != nil {
                return nil, fmt.Errorf("failed to open database: %w", err)
        }

        // Set connection pool settings
        db.SetMaxOpenConns(25)
        db.SetMaxIdleConns(5)
        db.SetConnMaxLifetime(5 * time.Minute)

        // Test the connection
        if err := db.Ping(); err != nil {
                return nil, fmt.Errorf("failed to ping database: %w", err)
        }

        return &PostgresDB{db}, nil
}

func (db *PostgresDB) Close() error {
        return db.DB.Close()
}

func (db *PostgresDB) HealthCheck() error {
        return db.Ping()
}

func (db *PostgresDB) RunMigrations() error {
        migrations := `
                CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

                CREATE TABLE IF NOT EXISTS users (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        email VARCHAR(255) UNIQUE NOT NULL,
                        name VARCHAR(255),
                        google_id VARCHAR(255) UNIQUE,
                        profile_image_url TEXT,
                        subscription_tier VARCHAR(50) DEFAULT 'free',
                        subscription_expires_at TIMESTAMP,
                        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS conversations (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                        title VARCHAR(255) NOT NULL,
                        language VARCHAR(10) DEFAULT 'ko',
                        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS messages (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
                        role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
                        content TEXT NOT NULL,
                        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS knowledge_items (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        title VARCHAR(255) NOT NULL,
                        content TEXT NOT NULL,
                        category VARCHAR(100),
                        tags TEXT[],
                        embedding_id VARCHAR(255),
                        source VARCHAR(255),
                        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS query_logs (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        user_id UUID REFERENCES users(id) ON DELETE SET NULL,
                        query TEXT NOT NULL,
                        response TEXT,
                        response_time_ms INTEGER,
                        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS subscription_plans (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        name VARCHAR(100) NOT NULL UNIQUE,
                        price_monthly DECIMAL(10, 2) NOT NULL,
                        price_yearly DECIMAL(10, 2),
                        max_queries_per_day INTEGER,
                        features JSONB,
                        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS user_subscriptions (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                        plan_id UUID NOT NULL REFERENCES subscription_plans(id),
                        stripe_subscription_id VARCHAR(255) UNIQUE,
                        status VARCHAR(50) NOT NULL,
                        current_period_start TIMESTAMP NOT NULL,
                        current_period_end TIMESTAMP NOT NULL,
                        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
                );

                INSERT INTO users (id, email, name, subscription_tier)
                VALUES ('00000000-0000-0000-0000-000000000000', 'anonymous@cleaners.ai', 'Anonymous User', 'free')
                ON CONFLICT (email) DO NOTHING;
        `

        _, err := db.Exec(migrations)
        if err != nil {
                return fmt.Errorf("failed to run migrations: %w", err)
        }

        return nil
}
