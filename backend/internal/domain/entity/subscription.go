package entity

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionPlan represents subscription plan type
type SubscriptionPlan string

const (
	PlanFree    SubscriptionPlan = "free"
	PlanMonthly SubscriptionPlan = "monthly"
	PlanYearly  SubscriptionPlan = "yearly"
)

// SubscriptionStatus represents subscription status
type SubscriptionStatus string

const (
	SubscriptionActive   SubscriptionStatus = "active"
	SubscriptionCanceled SubscriptionStatus = "canceled"
	SubscriptionExpired  SubscriptionStatus = "expired"
)

// Subscription represents a user subscription
type Subscription struct {
	ID               uuid.UUID          `json:"id"`
	UserID           uuid.UUID          `json:"user_id"`
	Plan             SubscriptionPlan   `json:"plan"`
	Status           SubscriptionStatus `json:"status"`
	StripeCustomerID string             `json:"stripe_customer_id"`
	StripeSubID      string             `json:"stripe_subscription_id"`
	CurrentPeriodEnd *time.Time         `json:"current_period_end,omitempty"`
	CanceledAt       *time.Time         `json:"canceled_at,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

// NewSubscription creates a new subscription
func NewSubscription(userID uuid.UUID, plan SubscriptionPlan) *Subscription {
	now := time.Now()
	return &Subscription{
		ID:        uuid.New(),
		UserID:    userID,
		Plan:      plan,
		Status:    SubscriptionActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsActive checks if subscription is active
func (s *Subscription) IsActive() bool {
	if s.Status != SubscriptionActive {
		return false
	}
	if s.CurrentPeriodEnd != nil && s.CurrentPeriodEnd.Before(time.Now()) {
		return false
	}
	return true
}

// IsPremium checks if subscription is premium (not free)
func (s *Subscription) IsPremium() bool {
	return s.Plan == PlanMonthly || s.Plan == PlanYearly
}

// Cancel cancels the subscription
func (s *Subscription) Cancel() {
	now := time.Now()
	s.Status = SubscriptionCanceled
	s.CanceledAt = &now
	s.UpdatedAt = now
}
