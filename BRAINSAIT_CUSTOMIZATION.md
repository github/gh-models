# BrainSait AI Platform - Customization Guide

> Transform `gh-models` into a branded, monetizable AI platform for targeted domains.

## ✅ Legal Status

**MIT License allows:**
- ✅ Commercial use
- ✅ Modification & rebranding
- ✅ Distribution
- ✅ Monetization
- ⚠️ Must include original license & copyright

---

## 🐳 1. Containerization

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o brainsait-models main.go

# Runtime stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/brainsait-models .
EXPOSE 8080
ENTRYPOINT ["./brainsait-models"]
```

### Docker Compose (with API wrapper)

```yaml
version: '3.8'
services:
  brainsait-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - GITHUB_TOKEN=${GITHUB_TOKEN}
      - BRAINSAIT_API_KEY=${BRAINSAIT_API_KEY}
    restart: unless-stopped
    
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: brainsait
      POSTGRES_USER: admin
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data

volumes:
  redis-data:
  postgres-data:
```

---

## 🏷️ 2. Branding Customization

### Files to Modify:

| File | Changes |
|------|---------|
| `cmd/root.go` | Update command names, descriptions |
| `internal/azuremodels/legal.go` | Update NOTICE |
| `go.mod` | Change module name |
| `README.md` | Complete rebrand |

### Example: cmd/root.go changes

```go
// Before
cmd := &cobra.Command{
    Use:   "models",
    Short: "GitHub Models extension",
    Long: heredoc.Docf(`
        GitHub Models CLI extension...
    `, "`"),
}

// After
cmd := &cobra.Command{
    Use:   "brainsait",
    Short: "BrainSait AI Platform",
    Long: heredoc.Docf(`
        BrainSait AI Platform - Enterprise AI for Healthcare, Arabic Markets & Developers.
        
        Powered by advanced language models with domain-specific optimizations.
    `, "`"),
}
```

---

## 🎯 3. Domain-Specific Customizations

### 3.1 Arabic Speakers Market 🌍

**Create: `prompts/arabic/`**

```yaml
# prompts/arabic/medical_ar.prompt.yml
name: "Arabic Medical Assistant"
model: "openai/gpt-4o"
messages:
  - role: system
    content: |
      أنت مساعد طبي متخصص باللغة العربية.
      قدم معلومات طبية دقيقة وواضحة.
      استخدم المصطلحات الطبية العربية الفصحى.
      
  - role: user
    content: "{{query}}"
testData:
  - query: "ما هي أعراض مرض السكري؟"
  - query: "كيف أتعامل مع ارتفاع ضغط الدم؟"
```

**Add RTL support in output:**

```go
// pkg/util/rtl.go
package util

import (
    "strings"
    "unicode"
)

func IsArabic(s string) bool {
    for _, r := range s {
        if unicode.Is(unicode.Arabic, r) {
            return true
        }
    }
    return false
}

func FormatRTL(s string) string {
    if IsArabic(s) {
        return "\u200F" + s + "\u200F" // RTL marks
    }
    return s
}
```

### 3.2 Healthcare & Insurance 🏥

**Create: `prompts/healthcare/`**

```yaml
# prompts/healthcare/insurance_claim.prompt.yml
name: "Insurance Claim Analyzer"
model: "openai/gpt-4o"
messages:
  - role: system
    content: |
      You are a healthcare insurance claims analyst.
      Analyze claims for:
      - Medical necessity
      - Coding accuracy (ICD-10, CPT)
      - Policy coverage alignment
      - Potential fraud indicators
      
      IMPORTANT: Never provide definitive coverage decisions.
      Always recommend human review for final determination.
      
  - role: user
    content: |
      Analyze this claim:
      {{claim_details}}
      
      Policy: {{policy_type}}
testData:
  - claim_details: "Emergency room visit, diagnosis: chest pain"
    policy_type: "PPO Standard"
```

**HIPAA Compliance additions:**

```go
// internal/compliance/hipaa.go
package compliance

import (
    "crypto/aes"
    "log"
)

type AuditLog struct {
    Timestamp   time.Time
    UserID      string
    Action      string
    DataAccessed string
    IPAddress   string
}

func LogPHIAccess(log AuditLog) error {
    // Implement HIPAA-compliant audit logging
}

func SanitizePHI(input string) string {
    // Remove/mask potential PHI before sending to external APIs
}
```

### 3.3 Developer Tools 💻

**Create: `prompts/developer/`**

```yaml
# prompts/developer/code_review.prompt.yml
name: "Code Review Assistant"
model: "openai/gpt-4o"
messages:
  - role: system
    content: |
      You are an expert code reviewer. Analyze code for:
      - Security vulnerabilities
      - Performance issues
      - Best practices adherence
      - Potential bugs
      
      Provide specific, actionable feedback with examples.
      
  - role: user
    content: |
      Review this {{language}} code:
      ```{{language}}
      {{code}}
      ```
testData:
  - language: "go"
    code: "func main() { fmt.Println(\"hello\") }"
```

---

## 💰 4. Monetization Architecture

### API Gateway Layer

```go
// cmd/api/main.go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/brainsait/platform/internal/billing"
    "github.com/brainsait/platform/internal/auth"
)

func main() {
    r := gin.Default()
    
    // Middleware
    r.Use(auth.APIKeyMiddleware())
    r.Use(billing.UsageTracker())
    r.Use(billing.RateLimiter())
    
    // Routes
    r.POST("/v1/chat", handlers.Chat)
    r.POST("/v1/eval", handlers.Eval)
    r.POST("/v1/generate", handlers.Generate)
    
    // Admin
    r.GET("/admin/usage", handlers.GetUsage)
    r.GET("/admin/billing", handlers.GetBilling)
    
    r.Run(":8080")
}
```

### Billing Service

```go
// internal/billing/tracker.go
package billing

type UsageTier struct {
    Name       string
    MonthlyCredits int
    PricePerCredit float64
    Features   []string
}

var Tiers = map[string]UsageTier{
    "free": {
        Name: "Free",
        MonthlyCredits: 100,
        PricePerCredit: 0,
        Features: []string{"basic_models", "community_support"},
    },
    "pro": {
        Name: "Professional",
        MonthlyCredits: 10000,
        PricePerCredit: 0.001,
        Features: []string{"all_models", "priority_support", "arabic_prompts"},
    },
    "enterprise": {
        Name: "Enterprise",
        MonthlyCredits: -1, // unlimited
        PricePerCredit: 0,
        Features: []string{"all_models", "dedicated_support", "custom_models", "hipaa", "on_premise"},
    },
}

func TrackUsage(userID string, credits int) error {
    // Track API usage against user's plan
}
```

### Stripe Integration

```go
// internal/billing/stripe.go
package billing

import (
    "github.com/stripe/stripe-go/v76"
    "github.com/stripe/stripe-go/v76/subscription"
)

func CreateSubscription(customerID, priceID string) (*stripe.Subscription, error) {
    params := &stripe.SubscriptionParams{
        Customer: stripe.String(customerID),
        Items: []*stripe.SubscriptionItemsParams{
            {Price: stripe.String(priceID)},
        },
    }
    return subscription.New(params)
}
```

---

## 📊 5. Database Schema

```sql
-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    api_key VARCHAR(64) UNIQUE NOT NULL,
    tier VARCHAR(50) DEFAULT 'free',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Usage tracking
CREATE TABLE usage_logs (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    endpoint VARCHAR(100),
    credits_used INT,
    model_used VARCHAR(100),
    domain VARCHAR(50), -- 'arabic', 'healthcare', 'developer'
    timestamp TIMESTAMP DEFAULT NOW()
);

-- Billing
CREATE TABLE billing_records (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    stripe_subscription_id VARCHAR(100),
    amount_cents INT,
    status VARCHAR(50),
    period_start DATE,
    period_end DATE
);
```

---

## 🚀 6. Quick Start

### Step 1: Fork & Clone
```bash
gh repo fork github/gh-models --clone=true
cd gh-models
git remote rename origin upstream
```

### Step 2: Rebrand
```bash
# Update module name
sed -i '' 's/github.com\/github\/gh-models/github.com\/brainsait\/platform/g' go.mod
find . -name "*.go" -exec sed -i '' 's/github.com\/github\/gh-models/github.com\/brainsait\/platform/g' {} \;
```

### Step 3: Build Container
```bash
docker build -t brainsait-platform:latest .
docker run -p 8080:8080 -e GITHUB_TOKEN=$GITHUB_TOKEN brainsait-platform:latest
```

### Step 4: Deploy
```bash
# Google Cloud Run
gcloud run deploy brainsait-api --source . --region=me-central1

# Or Kubernetes
kubectl apply -f k8s/deployment.yaml
```

---

## 📈 Revenue Projections

| Vertical | Target Market Size | Year 1 Goal |
|----------|-------------------|-------------|
| Arabic Speakers | 400M+ speakers | 1,000 users |
| Healthcare/Insurance | $4T industry | 50 enterprise clients |
| Developers | 27M+ worldwide | 5,000 users |

**Estimated ARR (Year 1):** $150K - $500K

---

## ⚠️ Important Considerations

1. **Model Access**: GitHub Models is in preview; consider alternative model providers for production
2. **Compliance**: Healthcare requires HIPAA BAA with model providers
3. **Data Residency**: Arabic market may require local data hosting
4. **Support**: Plan for Arabic-speaking support staff

---

## Next Steps

1. [ ] Fork repository
2. [ ] Create BrainSait organization on GitHub
3. [ ] Set up cloud infrastructure
4. [ ] Build MVP for one vertical first (recommend: Developers)
5. [ ] Validate with beta customers
6. [ ] Iterate and expand to other verticals
