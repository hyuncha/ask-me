package service

import (
	"cleaners-ai/pkg/vector"
	"strings"
)

// PartnerShop represents a partner cleaner shop
type PartnerShop struct {
	Name               string   `json:"name"`
	Zipcode            string   `json:"zipcode"`
	Priority           string   `json:"priority"`
	Rating             float64  `json:"rating"`
	Specialties        []string `json:"specialties"`
	SubscriptionStatus string   `json:"subscription_status"`
}

// PartnerShopService handles partner shop recommendations
type PartnerShopService struct {
	pineconeAssistant *vector.PineconeAssistantClient
}

// NewPartnerShopService creates a new partner shop service
func NewPartnerShopService(pineconeAssistant *vector.PineconeAssistantClient) *PartnerShopService {
	return &PartnerShopService{
		pineconeAssistant: pineconeAssistant,
	}
}

// ShouldRecommendShop determines if a professional cleaner should be recommended
func (s *PartnerShopService) ShouldRecommendShop(message string, successRate int) bool {
	message = strings.ToLower(message)

	// Condition 1: Success rate < 60%
	if successRate > 0 && successRate < 60 {
		return true
	}

	// Condition 2: Premium fabrics mentioned
	premiumFabrics := []string{
		"실크", "silk", "캐시미어", "cashmere", "가죽", "leather",
		"울", "wool", "린넨", "linen", "벨벳", "velvet",
		"스웨이드", "suede", "모피", "fur",
	}
	for _, fabric := range premiumFabrics {
		if strings.Contains(message, fabric) {
			return true
		}
	}

	// Condition 3: Time-sensitive stains (48+ hours mentioned)
	timePhrases := []string{
		"며칠", "일주일", "몇일", "오래", "48시간",
		"2일", "3일", "이틀", "사흘", "나흘",
	}
	for _, phrase := range timePhrases {
		if strings.Contains(message, phrase) {
			return true
		}
	}

	// Condition 4: Customer asking about professional service
	professionalPhrases := []string{
		"맡기", "세탁소", "전문", "드라이클리닝", "드라이 클리닝",
		"맡길까", "맡기면", "클리닝", "업체",
	}
	for _, phrase := range professionalPhrases {
		if strings.Contains(message, phrase) {
			return true
		}
	}

	return false
}

// GetPartnerShopsByLocation returns partner shops near the given zipcode
func (s *PartnerShopService) GetPartnerShopsByLocation(zipcode string) ([]PartnerShop, error) {
	// For now, return mock data
	// In production, this would query Pinecone Partner Cleaners index
	if zipcode == "" {
		return []PartnerShop{}, nil
	}

	// Mock partner shops for demonstration
	mockShops := []PartnerShop{
		{
			Name:               "클린마스터 세탁소",
			Zipcode:            zipcode,
			Priority:           "partner",
			Rating:             4.8,
			Specialties:        []string{"실크", "캐시미어", "명품가방"},
			SubscriptionStatus: "active",
		},
		{
			Name:               "프리미엄 드라이클리닝",
			Zipcode:            zipcode,
			Priority:           "partner",
			Rating:             4.6,
			Specialties:        []string{"정장", "웨딩드레스", "가죽"},
			SubscriptionStatus: "active",
		},
	}

	return mockShops, nil
}

// SearchPartnerShops searches for partner shops using Pinecone
func (s *PartnerShopService) SearchPartnerShops(query string, zipcode string) ([]PartnerShop, error) {
	if s.pineconeAssistant == nil {
		return s.GetPartnerShopsByLocation(zipcode)
	}

	// In production, this would:
	// 1. Query Pinecone Partner Cleaners index with the query and zipcode filter
	// 2. Return shops that match the criteria and have active subscriptions

	// For now, return location-based results
	return s.GetPartnerShopsByLocation(zipcode)
}
