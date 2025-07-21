package azuremodels

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRateLimitError(t *testing.T) {
	err := &RateLimitError{
		RetryAfter: 30 * time.Second,
		Message:    "Too many requests",
	}

	expected := "rate limited: Too many requests (retry after 30s)"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

func TestHandleHTTPError_RateLimit(t *testing.T) {
	client := &AzureClient{}

	tests := []struct {
		name               string
		statusCode         int
		headers            map[string]string
		expectedRetryAfter time.Duration
	}{
		{
			name:       "Rate limit with x-ratelimit-timeremaining header",
			statusCode: http.StatusTooManyRequests,
			headers: map[string]string{
				"x-ratelimit-timeremaining": "45",
			},
			expectedRetryAfter: 45 * time.Second,
		},
		{
			name:       "Rate limit with Retry-After header",
			statusCode: http.StatusTooManyRequests,
			headers: map[string]string{
				"Retry-After": "60",
			},
			expectedRetryAfter: 60 * time.Second,
		},
		{
			name:       "Rate limit with both headers - x-ratelimit-timeremaining takes precedence",
			statusCode: http.StatusTooManyRequests,
			headers: map[string]string{
				"x-ratelimit-timeremaining": "30",
				"Retry-After":               "90",
			},
			expectedRetryAfter: 30 * time.Second,
		},
		{
			name:               "Rate limit with no headers - default to 60s",
			statusCode:         http.StatusTooManyRequests,
			headers:            map[string]string{},
			expectedRetryAfter: 60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     make(http.Header),
				Body:       &mockReadCloser{reader: strings.NewReader("rate limit exceeded")},
			}

			for key, value := range tt.headers {
				resp.Header.Set(key, value)
			}

			err := client.handleHTTPError(resp)

			var rateLimitErr *RateLimitError
			if !isRateLimitError(err, &rateLimitErr) {
				t.Fatalf("Expected RateLimitError, got %T: %v", err, err)
			}

			if rateLimitErr.RetryAfter != tt.expectedRetryAfter {
				t.Errorf("Expected RetryAfter %v, got %v", tt.expectedRetryAfter, rateLimitErr.RetryAfter)
			}
		})
	}
}

// Helper function to check if error is a RateLimitError (mimics errors.As)
func isRateLimitError(err error, target **RateLimitError) bool {
	if rateLimitErr, ok := err.(*RateLimitError); ok {
		*target = rateLimitErr
		return true
	}
	return false
}

type mockReadCloser struct {
	reader *strings.Reader
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	return m.reader.Read(p)
}

func (m *mockReadCloser) Close() error {
	return nil
}
