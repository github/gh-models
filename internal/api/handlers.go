// Package api provides HTTP handlers for the BrainSait API
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/internal/middleware"
)

// API Response types
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *APIMeta    `json:"meta,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APIMeta struct {
	RequestID    string `json:"request_id,omitempty"`
	ProcessingMS int64  `json:"processing_ms,omitempty"`
	CreditsUsed  int    `json:"credits_used,omitempty"`
}

// Chat request/response types
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	ID      string      `json:"id"`
	Model   string      `json:"model"`
	Message ChatMessage `json:"message"`
	Usage   *ChatUsage  `json:"usage,omitempty"`
}

type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// =========================================
// Health & Info Handlers
// =========================================

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0",
		},
	})
}

func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"name":    "BrainSait AI Platform API",
			"version": "1.0.0",
			"docs":    "/docs",
			"endpoints": map[string]string{
				"chat":     "POST /v1/chat",
				"eval":     "POST /v1/eval",
				"generate": "POST /v1/generate",
				"models":   "GET /v1/models",
				"prompts":  "GET /v1/prompts",
			},
		},
	})
}

// =========================================
// Chat Handlers
// =========================================

func (s *Server) chatHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	// Validate request
	if req.Model == "" {
		s.errorResponse(w, http.StatusBadRequest, "MISSING_MODEL", "Model is required")
		return
	}
	if len(req.Messages) == 0 {
		s.errorResponse(w, http.StatusBadRequest, "MISSING_MESSAGES", "At least one message is required")
		return
	}

	// Get Azure client
	client, err := s.getAzureClient()
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, "CLIENT_ERROR", err.Error())
		return
	}

	// Convert messages
	messages := make([]azuremodels.ChatMessage, len(req.Messages))
	for i, msg := range req.Messages {
		content := msg.Content
		messages[i] = azuremodels.ChatMessage{
			Role:    azuremodels.ChatMessageRole(msg.Role),
			Content: &content,
		}
	}

	// Create completion options
	opts := azuremodels.ChatCompletionOptions{
		Messages: messages,
		Model:    req.Model,
	}
	if req.MaxTokens > 0 {
		opts.MaxTokens = &req.MaxTokens
	}
	if req.Temperature > 0 {
		opts.Temperature = &req.Temperature
	}

	// Call Azure Models API
	response, err := client.GetChatCompletionStream(r.Context(), opts, "")
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, "API_ERROR", err.Error())
		return
	}

	// Read the first (and possibly only) response from the stream
	completion, err := response.Reader.Read()
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, "API_ERROR", err.Error())
		return
	}

	// Extract response
	var content string
	if len(completion.Choices) > 0 && completion.Choices[0].Message != nil {
		content = *completion.Choices[0].Message.Content
	}

	processingMs := time.Since(start).Milliseconds()

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data: ChatResponse{
			ID:    fmt.Sprintf("chat-%d", time.Now().UnixNano()),
			Model: req.Model,
			Message: ChatMessage{
				Role:    "assistant",
				Content: content,
			},
		},
		Meta: &APIMeta{
			RequestID:    r.Header.Get("X-Request-ID"),
			ProcessingMS: processingMs,
			CreditsUsed:  1,
		},
	})
}

func (s *Server) chatStreamHandler(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"error\": \"Invalid request\"}\n\n")
		return
	}

	client, err := s.getAzureClient()
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"error\": \"%s\"}\n\n", err.Error())
		return
	}

	// Convert messages
	messages := make([]azuremodels.ChatMessage, len(req.Messages))
	for i, msg := range req.Messages {
		content := msg.Content
		messages[i] = azuremodels.ChatMessage{
			Role:    azuremodels.ChatMessageRole(msg.Role),
			Content: &content,
		}
	}

	opts := azuremodels.ChatCompletionOptions{
		Messages: messages,
		Model:    req.Model,
		Stream:   true,
	}

	// Stream response
	response, err := client.GetChatCompletionStream(r.Context(), opts, "")
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"error\": \"%s\"}\n\n", err.Error())
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Fprintf(w, "event: error\ndata: {\"error\": \"Streaming not supported\"}\n\n")
		return
	}

	for {
		event, err := response.Reader.Read()
		if err != nil {
			break
		}

		if len(event.Choices) > 0 {
			delta := event.Choices[0].Delta
			if delta != nil && delta.Content != nil {
				data, _ := json.Marshal(map[string]string{"content": *delta.Content})
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	}

	fmt.Fprintf(w, "event: done\ndata: {}\n\n")
	flusher.Flush()
}

// =========================================
// Model Handlers
// =========================================

func (s *Server) listModelsHandler(w http.ResponseWriter, r *http.Request) {
	client, err := s.getAzureClient()
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, "CLIENT_ERROR", err.Error())
		return
	}

	models, err := client.ListModels(r.Context())
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, "API_ERROR", err.Error())
		return
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    models,
	})
}

func (s *Server) viewModelHandler(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("modelId")
	if modelID == "" {
		s.errorResponse(w, http.StatusBadRequest, "MISSING_MODEL_ID", "Model ID is required")
		return
	}

	// URL decode the model ID (e.g., openai%2Fgpt-4o -> openai/gpt-4o)
	modelID = strings.ReplaceAll(modelID, "%2F", "/")

	client, err := s.getAzureClient()
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, "CLIENT_ERROR", err.Error())
		return
	}

	// First list models to find the one we're looking for
	models, err := client.ListModels(r.Context())
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, "API_ERROR", err.Error())
		return
	}

	// Find the model by name
	var modelSummary *azuremodels.ModelSummary
	for _, model := range models {
		if model.HasName(modelID) {
			modelSummary = model
			break
		}
	}

	if modelSummary == nil {
		s.errorResponse(w, http.StatusNotFound, "MODEL_NOT_FOUND", fmt.Sprintf("Model %s not found", modelID))
		return
	}

	// Get detailed model info
	model, err := client.GetModelDetails(r.Context(), modelSummary.Registry, modelSummary.Name, modelSummary.Version)
	if err != nil {
		s.errorResponse(w, http.StatusNotFound, "MODEL_NOT_FOUND", err.Error())
		return
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    model,
	})
}

// =========================================
// Eval Handler
// =========================================

func (s *Server) evalHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement eval logic using the eval package
	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]string{
			"message": "Evaluation endpoint - implementation pending",
		},
	})
}

// =========================================
// Generate Handler
// =========================================

func (s *Server) generateHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement generate logic using the generate package
	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]string{
			"message": "Generate endpoint - implementation pending",
		},
	})
}

// =========================================
// Prompt Library Handlers
// =========================================

func (s *Server) listPromptsHandler(w http.ResponseWriter, r *http.Request) {
	// Return available prompt templates
	prompts := []map[string]interface{}{
		{
			"id":          "arabic-general",
			"name":        "Arabic General Assistant",
			"domain":      "arabic",
			"description": "مساعد ذكاء اصطناعي عام باللغة العربية",
		},
		{
			"id":          "arabic-medical",
			"name":        "Arabic Medical Advisor",
			"domain":      "arabic",
			"description": "مساعد طبي للمعلومات الصحية العامة",
		},
		{
			"id":          "healthcare-claims",
			"name":        "Insurance Claim Analyzer",
			"domain":      "healthcare",
			"description": "Analyzes healthcare insurance claims",
		},
		{
			"id":          "healthcare-pa",
			"name":        "Prior Authorization Assistant",
			"domain":      "healthcare",
			"description": "Assists with prior authorization requests",
		},
		{
			"id":          "developer-review",
			"name":        "Code Review Assistant",
			"domain":      "developer",
			"description": "Expert code review and analysis",
		},
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    prompts,
	})
}

func (s *Server) listPromptsByDomainHandler(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	// Filter prompts by domain
	// TODO: Implement database query

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]string{
			"domain":  domain,
			"message": "Domain-specific prompts",
		},
	})
}

func (s *Server) executePromptHandler(w http.ResponseWriter, r *http.Request) {
	// Execute a prompt template with provided variables
	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]string{
			"message": "Prompt execution endpoint",
		},
	})
}

// =========================================
// User Handlers
// =========================================

func (s *Server) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey)

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"user_id": userID,
			"tier":    "pro",
			"domain":  "developer",
		},
	})
}

func (s *Server) getUsageHandler(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"period_start":      time.Now().Format("2006-01-01"),
			"period_end":        time.Now().AddDate(0, 1, 0).Format("2006-01-01"),
			"credits_used":      150,
			"credits_remaining": 9850,
			"total_requests":    245,
		},
	})
}

// =========================================
// Billing Handlers
// =========================================

func (s *Server) createCheckoutHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PriceID    string `json:"price_id"`
		SuccessURL string `json:"success_url"`
		CancelURL  string `json:"cancel_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request")
		return
	}

	session, err := s.billing.CreateCheckoutSession(req.PriceID, req.SuccessURL, req.CancelURL)
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, "STRIPE_ERROR", err.Error())
		return
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]string{
			"checkout_url": session.URL,
		},
	})
}

func (s *Server) createPortalHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID string `json:"customer_id"`
		ReturnURL  string `json:"return_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request")
		return
	}

	session, err := s.billing.CreatePortalSession(req.CustomerID, req.ReturnURL)
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, "STRIPE_ERROR", err.Error())
		return
	}

	s.jsonResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]string{
			"portal_url": session.URL,
		},
	})
}

func (s *Server) stripeWebhookHandler(w http.ResponseWriter, r *http.Request) {
	// Handle Stripe webhook events
	// TODO: Implement webhook signature verification and event handling
	w.WriteHeader(http.StatusOK)
}

// =========================================
// Helper Methods
// =========================================

func (s *Server) getAzureClient() (azuremodels.Client, error) {
	token := s.config.GitHubToken
	if token == "" {
		token, _ = auth.TokenForHost("github.com")
	}
	if token == "" {
		return nil, fmt.Errorf("no GitHub token available")
	}
	return azuremodels.NewDefaultAzureClient(token)
}

func (s *Server) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) errorResponse(w http.ResponseWriter, status int, code, message string) {
	s.jsonResponse(w, status, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}
