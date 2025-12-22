# BrainSait API Integration Guide

Complete guide for integrating with the BrainSait AI Platform API.

## Quick Start

### Base URL

```
Production: https://api.brainsait.ai/v1
Staging:    https://staging-api.brainsait.ai/v1
Local:      http://localhost:8080/v1
```

### Authentication

All API requests require authentication via API key:

```bash
curl -H "X-API-Key: YOUR_API_KEY" https://api.brainsait.ai/v1/models
```

Or using Bearer token:

```bash
curl -H "Authorization: Bearer YOUR_API_KEY" https://api.brainsait.ai/v1/models
```

### Get Your API Key

1. Sign up at [brainsait.ai](https://brainsait.ai)
2. Navigate to Dashboard → API Keys
3. Create a new API key
4. Store securely - keys are only shown once

## API Endpoints

### Health Check

```http
GET /health
```

No authentication required.

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2025-01-01T00:00:00Z",
    "version": "1.0.0"
  }
}
```

### Chat Completion

```http
POST /v1/chat
```

Generate AI responses for conversational messages.

**Request:**
```json
{
  "model": "openai/gpt-4o",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello, how are you?"}
  ],
  "max_tokens": 1000,
  "temperature": 0.7
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "chat-123456",
    "model": "openai/gpt-4o",
    "message": {
      "role": "assistant",
      "content": "Hello! I'm doing well, thank you for asking..."
    }
  },
  "meta": {
    "request_id": "req-abc123",
    "processing_ms": 1250,
    "credits_used": 1
  }
}
```

### Streaming Chat

```http
POST /v1/chat/stream
```

Stream responses using Server-Sent Events (SSE).

**Request:** Same as `/v1/chat`

**Response:** SSE stream
```
data: {"content": "Hello"}
data: {"content": "!"}
data: {"content": " I'm"}
data: {"content": " doing"}
...
event: done
data: {}
```

**Example (curl):**
```bash
curl -N -X POST "https://api.brainsait.ai/v1/chat/stream" \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openai/gpt-4o",
    "messages": [{"role": "user", "content": "Tell me a story"}]
  }'
```

### List Models

```http
GET /v1/models
```

List all available AI models.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "openai/gpt-4o",
      "name": "GPT-4o",
      "publisher": "OpenAI",
      "task": "chat-completion"
    },
    {
      "id": "openai/gpt-4o-mini",
      "name": "GPT-4o mini",
      "publisher": "OpenAI",
      "task": "chat-completion"
    }
  ]
}
```

### Get Model Details

```http
GET /v1/models/{model_id}
```

Get detailed information about a specific model.

**Example:**
```bash
curl -H "X-API-Key: YOUR_API_KEY" \
  "https://api.brainsait.ai/v1/models/openai%2Fgpt-4o"
```

### Prompt Library

```http
GET /v1/prompts
```

List available prompt templates.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "arabic-general",
      "name": "Arabic General Assistant",
      "domain": "arabic",
      "description": "مساعد ذكاء اصطناعي عام باللغة العربية"
    },
    {
      "id": "healthcare-claims",
      "name": "Insurance Claim Analyzer",
      "domain": "healthcare",
      "description": "Analyzes healthcare insurance claims"
    }
  ]
}
```

### Prompts by Domain

```http
GET /v1/prompts/{domain}
```

Filter prompts by domain (arabic, healthcare, developer).

### Execute Prompt

```http
POST /v1/prompts/execute
```

Execute a prompt template with variables.

**Request:**
```json
{
  "prompt_id": "healthcare-claims",
  "variables": {
    "claim_data": "Patient: John Doe, Procedure: MRI..."
  }
}
```

### User Profile

```http
GET /v1/user
```

Get current user information.

### Usage Statistics

```http
GET /v1/user/usage
```

Get API usage statistics for current billing period.

**Response:**
```json
{
  "success": true,
  "data": {
    "period_start": "2025-01-01",
    "period_end": "2025-02-01",
    "credits_used": 150,
    "credits_remaining": 9850,
    "total_requests": 245
  }
}
```

## Billing Endpoints

### Create Checkout Session

```http
POST /v1/billing/checkout
```

Create a Stripe checkout session for subscription.

**Request:**
```json
{
  "price_id": "price_developer_monthly",
  "success_url": "https://yourapp.com/success",
  "cancel_url": "https://yourapp.com/cancel"
}
```

### Customer Portal

```http
POST /v1/billing/portal
```

Create a Stripe billing portal session.

## SDKs & Code Examples

### Python

```python
import requests

class BrainSaitClient:
    def __init__(self, api_key: str, base_url: str = "https://api.brainsait.ai/v1"):
        self.api_key = api_key
        self.base_url = base_url
        self.headers = {"X-API-Key": api_key}
    
    def chat(self, model: str, messages: list, **kwargs) -> dict:
        response = requests.post(
            f"{self.base_url}/chat",
            headers=self.headers,
            json={"model": model, "messages": messages, **kwargs}
        )
        response.raise_for_status()
        return response.json()
    
    def list_models(self) -> list:
        response = requests.get(f"{self.base_url}/models", headers=self.headers)
        response.raise_for_status()
        return response.json()["data"]

# Usage
client = BrainSaitClient("your-api-key")

# Chat completion
result = client.chat(
    model="openai/gpt-4o",
    messages=[{"role": "user", "content": "Hello!"}]
)
print(result["data"]["message"]["content"])
```

### JavaScript/TypeScript

```typescript
class BrainSaitClient {
  private apiKey: string;
  private baseUrl: string;

  constructor(apiKey: string, baseUrl = "https://api.brainsait.ai/v1") {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async chat(model: string, messages: Array<{role: string; content: string}>) {
    const response = await fetch(`${this.baseUrl}/chat`, {
      method: "POST",
      headers: {
        "X-API-Key": this.apiKey,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ model, messages }),
    });
    return response.json();
  }

  async listModels() {
    const response = await fetch(`${this.baseUrl}/models`, {
      headers: { "X-API-Key": this.apiKey },
    });
    const data = await response.json();
    return data.data;
  }

  // Streaming with async generator
  async *chatStream(model: string, messages: Array<{role: string; content: string}>) {
    const response = await fetch(`${this.baseUrl}/chat/stream`, {
      method: "POST",
      headers: {
        "X-API-Key": this.apiKey,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ model, messages }),
    });

    const reader = response.body!.getReader();
    const decoder = new TextDecoder();

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      
      const text = decoder.decode(value);
      for (const line of text.split('\n')) {
        if (line.startsWith('data: ')) {
          const data = JSON.parse(line.slice(6));
          if (data.content) yield data.content;
        }
      }
    }
  }
}

// Usage
const client = new BrainSaitClient("your-api-key");

// Chat completion
const result = await client.chat("openai/gpt-4o", [
  { role: "user", content: "Hello!" }
]);
console.log(result.data.message.content);

// Streaming
for await (const chunk of client.chatStream("openai/gpt-4o", messages)) {
  process.stdout.write(chunk);
}
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type BrainSaitClient struct {
    APIKey  string
    BaseURL string
}

func NewClient(apiKey string) *BrainSaitClient {
    return &BrainSaitClient{
        APIKey:  apiKey,
        BaseURL: "https://api.brainsait.ai/v1",
    }
}

func (c *BrainSaitClient) Chat(model string, messages []map[string]string) (map[string]interface{}, error) {
    body, _ := json.Marshal(map[string]interface{}{
        "model":    model,
        "messages": messages,
    })

    req, _ := http.NewRequest("POST", c.BaseURL+"/chat", bytes.NewBuffer(body))
    req.Header.Set("X-API-Key", c.APIKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    return result, nil
}

func main() {
    client := NewClient("your-api-key")
    
    result, _ := client.Chat("openai/gpt-4o", []map[string]string{
        {"role": "user", "content": "Hello!"},
    })
    
    fmt.Println(result)
}
```

### cURL Examples

```bash
# Health check
curl https://api.brainsait.ai/health

# List models
curl -H "X-API-Key: YOUR_KEY" https://api.brainsait.ai/v1/models

# Chat completion
curl -X POST https://api.brainsait.ai/v1/chat \
  -H "X-API-Key: YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openai/gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# Arabic assistant
curl -X POST https://api.brainsait.ai/v1/chat \
  -H "X-API-Key: YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openai/gpt-4o",
    "messages": [
      {"role": "system", "content": "أنت مساعد ذكاء اصطناعي يتحدث العربية بطلاقة"},
      {"role": "user", "content": "مرحباً، كيف حالك؟"}
    ]
  }'
```

## Error Handling

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Missing or invalid API key |
| `RATE_LIMITED` | 429 | Too many requests |
| `INVALID_JSON` | 400 | Malformed request body |
| `MISSING_MODEL` | 400 | Model parameter required |
| `MISSING_MESSAGES` | 400 | Messages parameter required |
| `MODEL_NOT_FOUND` | 404 | Specified model doesn't exist |
| `INSUFFICIENT_CREDITS` | 402 | Account has no remaining credits |
| `API_ERROR` | 500 | Internal server error |

### Rate Limits

| Tier | Requests/min | Requests/day |
|------|--------------|--------------|
| Free | 10 | 100 |
| Developer | 60 | 10,000 |
| Team | 120 | 50,000 |
| Enterprise | Custom | Custom |

## Webhooks

Configure webhooks for billing events:

```http
POST /webhooks/stripe
```

Supported events:
- `checkout.session.completed` - New subscription
- `invoice.paid` - Successful payment
- `customer.subscription.updated` - Plan change
- `customer.subscription.deleted` - Cancellation

## Support

- **Documentation**: https://docs.brainsait.ai
- **Status Page**: https://status.brainsait.ai
- **Support Email**: support@brainsait.ai
- **Developer Discord**: https://discord.gg/brainsait

## Changelog

### v1.0.0 (2025-01-01)
- Initial API release
- Chat completions with streaming
- Model listing and details
- Prompt library with domain filters
- Stripe billing integration
