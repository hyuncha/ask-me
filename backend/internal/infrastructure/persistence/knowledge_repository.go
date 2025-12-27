package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"cleaners-ai/internal/domain/entity"
)

type KnowledgeRepository struct {
	db *sql.DB
}

func NewKnowledgeRepository(db *sql.DB) *KnowledgeRepository {
	return &KnowledgeRepository{db: db}
}

func (r *KnowledgeRepository) Create(ctx context.Context, knowledge *entity.Knowledge) error {
	tagsJSON, err := json.Marshal(knowledge.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		INSERT INTO knowledge_items (id, title, content, category, difficulty, tags, language, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.ExecContext(ctx, query,
		knowledge.ID,
		knowledge.Title,
		knowledge.Content,
		knowledge.Category,
		knowledge.Difficulty,
		tagsJSON,
		knowledge.Language,
		knowledge.Status,
		knowledge.CreatedBy,
		knowledge.CreatedAt,
		knowledge.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create knowledge: %w", err)
	}

	return nil
}

func (r *KnowledgeRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Knowledge, error) {
	query := `
		SELECT id, title, content, category, difficulty, tags, language, status, created_by, created_at, updated_at
		FROM knowledge_items
		WHERE id = $1
	`

	var knowledge entity.Knowledge
	var tagsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&knowledge.ID,
		&knowledge.Title,
		&knowledge.Content,
		&knowledge.Category,
		&knowledge.Difficulty,
		&tagsJSON,
		&knowledge.Language,
		&knowledge.Status,
		&knowledge.CreatedBy,
		&knowledge.CreatedAt,
		&knowledge.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("knowledge not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge: %w", err)
	}

	if err := json.Unmarshal(tagsJSON, &knowledge.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	return &knowledge, nil
}

func (r *KnowledgeRepository) GetByCategory(ctx context.Context, category entity.KnowledgeCategory) ([]*entity.Knowledge, error) {
	query := `
		SELECT id, title, content, category, difficulty, tags, language, status, created_by, created_at, updated_at
		FROM knowledge_items
		WHERE category = $1 AND status = 'active'
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to query knowledge: %w", err)
	}
	defer rows.Close()

	var items []*entity.Knowledge
	for rows.Next() {
		var knowledge entity.Knowledge
		var tagsJSON []byte

		if err := rows.Scan(
			&knowledge.ID,
			&knowledge.Title,
			&knowledge.Content,
			&knowledge.Category,
			&knowledge.Difficulty,
			&tagsJSON,
			&knowledge.Language,
			&knowledge.Status,
			&knowledge.CreatedBy,
			&knowledge.CreatedAt,
			&knowledge.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan knowledge: %w", err)
		}

		if err := json.Unmarshal(tagsJSON, &knowledge.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		items = append(items, &knowledge)
	}

	return items, nil
}

func (r *KnowledgeRepository) GetAll(ctx context.Context) ([]*entity.Knowledge, error) {
	query := `
		SELECT id, title, content, category, difficulty, tags, language, status, created_by, created_at, updated_at
		FROM knowledge_items
		WHERE status = 'active'
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query knowledge: %w", err)
	}
	defer rows.Close()

	var items []*entity.Knowledge
	for rows.Next() {
		var knowledge entity.Knowledge
		var tagsJSON []byte

		if err := rows.Scan(
			&knowledge.ID,
			&knowledge.Title,
			&knowledge.Content,
			&knowledge.Category,
			&knowledge.Difficulty,
			&tagsJSON,
			&knowledge.Language,
			&knowledge.Status,
			&knowledge.CreatedBy,
			&knowledge.CreatedAt,
			&knowledge.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan knowledge: %w", err)
		}

		if err := json.Unmarshal(tagsJSON, &knowledge.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		items = append(items, &knowledge)
	}

	return items, nil
}

func (r *KnowledgeRepository) Update(ctx context.Context, knowledge *entity.Knowledge) error {
	tagsJSON, err := json.Marshal(knowledge.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		UPDATE knowledge_items
		SET title = $2, content = $3, category = $4, difficulty = $5, tags = $6, language = $7, status = $8, updated_at = $9
		WHERE id = $1
	`

	_, err = r.db.ExecContext(ctx, query,
		knowledge.ID,
		knowledge.Title,
		knowledge.Content,
		knowledge.Category,
		knowledge.Difficulty,
		tagsJSON,
		knowledge.Language,
		knowledge.Status,
		knowledge.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update knowledge: %w", err)
	}

	return nil
}

func (r *KnowledgeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM knowledge_items WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete knowledge: %w", err)
	}

	return nil
}
