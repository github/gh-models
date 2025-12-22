// Package api provides the HTTP API server implementation
package api

import (
	"net/http"
	"os"

	"github.com/github/gh-models/internal/billing"
	"github.com/github/gh-models/internal/middleware"
)

// Config holds the API server configuration
type Config struct {
	Port                string
	Environment         string
	GitHubToken         string
	StripeSecretKey     string
	StripeWebhookSecret string
	JWTSecret           string
	DatabaseURL         string
	RedisURL            string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Port:                getEnv("PORT", "8080"),
		Environment:         getEnv("ENVIRONMENT", "development"),
		GitHubToken:         os.Getenv("GITHUB_TOKEN"),
		StripeSecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		RedisURL:            getEnv("REDIS_URL", "redis://localhost:6379"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Server represents the API server
type Server struct {
	config  *Config
	billing *billing.Service
}

// NewServer creates a new API server instance
func NewServer(cfg *Config, billingService *billing.Service) *Server {
	return &Server{
		config:  cfg,
		billing: billingService,
	}
}

// SetupRoutes configures all API routes
func (s *Server) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Health check (no auth)
	mux.HandleFunc("GET /health", s.healthHandler)
	mux.HandleFunc("GET /", s.rootHandler)

	// API v1 routes
	mux.HandleFunc("POST /v1/chat", s.withMiddleware(s.chatHandler))
	mux.HandleFunc("POST /v1/chat/stream", s.withMiddleware(s.chatStreamHandler))
	mux.HandleFunc("POST /v1/eval", s.withMiddleware(s.evalHandler))
	mux.HandleFunc("POST /v1/generate", s.withMiddleware(s.generateHandler))
	mux.HandleFunc("GET /v1/models", s.withMiddleware(s.listModelsHandler))
	mux.HandleFunc("GET /v1/models/{modelId}", s.withMiddleware(s.viewModelHandler))

	// Prompt library
	mux.HandleFunc("GET /v1/prompts", s.withMiddleware(s.listPromptsHandler))
	mux.HandleFunc("GET /v1/prompts/{domain}", s.withMiddleware(s.listPromptsByDomainHandler))
	mux.HandleFunc("POST /v1/prompts/execute", s.withMiddleware(s.executePromptHandler))

	// User management
	mux.HandleFunc("GET /v1/user", s.withMiddleware(s.getUserHandler))
	mux.HandleFunc("GET /v1/user/usage", s.withMiddleware(s.getUsageHandler))

	// Billing (Stripe)
	mux.HandleFunc("POST /v1/billing/checkout", s.withMiddleware(s.createCheckoutHandler))
	mux.HandleFunc("POST /v1/billing/portal", s.withMiddleware(s.createPortalHandler))
	mux.HandleFunc("POST /webhooks/stripe", s.stripeWebhookHandler)

	// Apply global middleware
	handler := middleware.CORS(mux)
	handler = middleware.Logger(handler)
	handler = middleware.Recovery(handler)

	return handler
}

// withMiddleware applies authentication and rate limiting
func (s *Server) withMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Apply auth middleware
		middleware.APIKeyAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Apply rate limiting
			middleware.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Apply usage tracking
				middleware.UsageTracker(http.HandlerFunc(next)).ServeHTTP(w, r)
			})).ServeHTTP(w, r)
		})).ServeHTTP(w, r)
	}
}
