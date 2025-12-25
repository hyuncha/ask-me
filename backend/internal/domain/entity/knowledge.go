package entity

import (
	"time"

	"github.com/google/uuid"
)

// KnowledgeCategory represents the 7 fixed categories
type KnowledgeCategory string

const (
	CategoryStainRemoval       KnowledgeCategory = "stain_removal"        // 얼룩제거
	CategoryFabricUnderstanding KnowledgeCategory = "fabric_understanding" // 원단이해
	CategoryAccidentPrevention KnowledgeCategory = "accident_prevention"  // 세탁사고와 방지
	CategoryLaundryTechnique   KnowledgeCategory = "laundry_technique"    // 세탁기술
	CategoryEquipmentOperation KnowledgeCategory = "equipment_operation"  // 세탁장비운영
	CategoryMarketing          KnowledgeCategory = "marketing"            // 마케팅과 고객관리
	CategoryOthers             KnowledgeCategory = "others"               // 기타
)

// KnowledgeDifficulty represents knowledge difficulty level
type KnowledgeDifficulty string

const (
	DifficultyBasic  KnowledgeDifficulty = "basic"
	DifficultyExpert KnowledgeDifficulty = "expert"
)

// KnowledgeStatus represents knowledge document status
type KnowledgeStatus string

const (
	StatusActive   KnowledgeStatus = "active"
	StatusInactive KnowledgeStatus = "inactive"
	StatusDraft    KnowledgeStatus = "draft"
)

// Knowledge represents a knowledge document
type Knowledge struct {
	ID         uuid.UUID             `json:"id"`
	Title      string                `json:"title"`
	Content    string                `json:"content"`
	Category   KnowledgeCategory     `json:"category"`
	Difficulty KnowledgeDifficulty   `json:"difficulty"`
	Tags       []string              `json:"tags"`
	Language   string                `json:"language"` // "KR" or "EN"
	Status     KnowledgeStatus       `json:"status"`
	CreatedBy  uuid.UUID             `json:"created_by"`
	CreatedAt  time.Time             `json:"created_at"`
	UpdatedAt  time.Time             `json:"updated_at"`
}

// NewKnowledge creates a new knowledge document
func NewKnowledge(title, content string, category KnowledgeCategory, difficulty KnowledgeDifficulty, tags []string, language string, createdBy uuid.UUID) *Knowledge {
	now := time.Now()
	return &Knowledge{
		ID:         uuid.New(),
		Title:      title,
		Content:    content,
		Category:   category,
		Difficulty: difficulty,
		Tags:       tags,
		Language:   language,
		Status:     StatusActive, // Default to active
		CreatedBy:  createdBy,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// IsActive checks if knowledge is active
func (k *Knowledge) IsActive() bool {
	return k.Status == StatusActive
}

// Activate activates the knowledge document
func (k *Knowledge) Activate() {
	k.Status = StatusActive
	k.UpdatedAt = time.Now()
}

// Deactivate deactivates the knowledge document
func (k *Knowledge) Deactivate() {
	k.Status = StatusInactive
	k.UpdatedAt = time.Now()
}
