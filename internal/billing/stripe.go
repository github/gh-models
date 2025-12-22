// Package billing provides Stripe integration for BrainSait
package billing

import (
	"fmt"
	"os"
)

// Service handles Stripe payment operations
type Service struct {
	secretKey     string
	webhookSecret string
}

// NewService creates a new Billing Service
func NewService(secretKey, webhookSecret string) *Service {
	if secretKey == "" {
		secretKey = os.Getenv("STRIPE_SECRET_KEY")
	}
	if webhookSecret == "" {
		webhookSecret = os.Getenv("STRIPE_WEBHOOK_SECRET")
	}

	return &Service{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
	}
}

// Pricing tiers for BrainSait
type PricingTier struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	PriceID     string   `json:"price_id,omitempty"`
	Amount      int64    `json:"amount"`
	Currency    string   `json:"currency"`
	Interval    string   `json:"interval,omitempty"`
	Features    []string `json:"features"`
}

// GetPricingTiers returns available pricing plans
func (s *Service) GetPricingTiers() []PricingTier {
	return []PricingTier{
		{
			ID:          "free",
			Name:        "Free",
			Description: "Get started with basic features",
			Amount:      0,
			Currency:    "usd",
			Features: []string{
				"100 API calls/month",
				"Community support",
				"Basic models access",
			},
		},
		{
			ID:          "developer",
			Name:        "Developer",
			Description: "For individual developers",
			PriceID:     os.Getenv("STRIPE_PRICE_DEVELOPER"),
			Amount:      2900, // $29.00
			Currency:    "usd",
			Interval:    "month",
			Features: []string{
				"10,000 API calls/month",
				"Email support",
				"All models access",
				"Prompt evaluation",
				"Basic analytics",
			},
		},
		{
			ID:          "team",
			Name:        "Team",
			Description: "For small teams",
			PriceID:     os.Getenv("STRIPE_PRICE_TEAM"),
			Amount:      9900, // $99.00
			Currency:    "usd",
			Interval:    "month",
			Features: []string{
				"50,000 API calls/month",
				"Priority support",
				"All models access",
				"Prompt evaluation",
				"Test generation",
				"Advanced analytics",
				"Team management",
			},
		},
		{
			ID:          "enterprise",
			Name:        "Enterprise",
			Description: "For large organizations",
			PriceID:     os.Getenv("STRIPE_PRICE_ENTERPRISE"),
			Amount:      49900, // $499.00
			Currency:    "usd",
			Interval:    "month",
			Features: []string{
				"Unlimited API calls",
				"24/7 support",
				"All models access",
				"Prompt evaluation",
				"Test generation",
				"Advanced analytics",
				"Team management",
				"Custom integrations",
				"SLA guarantee",
				"Dedicated success manager",
			},
		},
	}
}

// CheckoutSession represents a Stripe checkout session
type CheckoutSession struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// PortalSession represents a Stripe billing portal session
type PortalSession struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// CreateCheckoutSession creates a Stripe checkout session
// This is a stub - actual Stripe integration requires the stripe-go package
func (s *Service) CreateCheckoutSession(priceID, successURL, cancelURL string) (*CheckoutSession, error) {
	if s.secretKey == "" {
		return nil, fmt.Errorf("Stripe secret key not configured")
	}

	// TODO: Implement actual Stripe checkout session creation
	// This requires: go get github.com/stripe/stripe-go/v78

	return &CheckoutSession{
		ID:  "cs_test_placeholder",
		URL: successURL + "?session_id=cs_test_placeholder",
	}, nil
}

// CreatePortalSession creates a Stripe billing portal session
func (s *Service) CreatePortalSession(customerID, returnURL string) (*PortalSession, error) {
	if s.secretKey == "" {
		return nil, fmt.Errorf("Stripe secret key not configured")
	}

	// TODO: Implement actual Stripe portal session creation

	return &PortalSession{
		ID:  "bps_test_placeholder",
		URL: returnURL + "?portal_session=bps_test_placeholder",
	}, nil
}

// HandleWebhook processes Stripe webhook events
func (s *Service) HandleWebhook(payload []byte, signature string) error {
	// TODO: Implement webhook verification and handling
	return nil
}
