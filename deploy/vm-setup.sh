#!/bin/bash
# VM Setup Script for BrainSait Production Deployment
# Designed for Ubuntu 22.04 LTS

set -e

echo "🖥️  BrainSait VM Setup"
echo "======================"

# Configuration
BRAINSAIT_USER="${BRAINSAIT_USER:-brainsait}"
DOCKER_USER="${DOCKER_USER:-brainsait}"
IMAGE_NAME="${IMAGE_NAME:-brainsait-ai}"
DOMAIN="${DOMAIN:-brainsait.com}"

# Update system
echo ""
echo "📦 Step 1: System Update"
sudo apt update && sudo apt upgrade -y

# Install Docker
echo ""
echo "📦 Step 2: Install Docker"
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    rm get-docker.sh
    
    # Add current user to docker group
    sudo usermod -aG docker $USER
    echo "⚠️  You may need to log out and back in for docker group membership"
fi

# Install Docker Compose
echo ""
echo "📦 Step 3: Install Docker Compose"
if ! command -v docker-compose &> /dev/null; then
    sudo apt install -y docker-compose-plugin
fi

# Install cloudflared
echo ""
echo "📦 Step 4: Install Cloudflared"
if ! command -v cloudflared &> /dev/null; then
    curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
    sudo dpkg -i cloudflared.deb
    rm cloudflared.deb
fi

# Create application directory
echo ""
echo "📁 Step 5: Create Application Directory"
sudo mkdir -p /opt/brainsait
sudo chown $USER:$USER /opt/brainsait
cd /opt/brainsait

# Create docker-compose file
echo ""
echo "📝 Step 6: Create Docker Compose Configuration"
cat > docker-compose.yml << 'DOCKER_COMPOSE'
version: '3.8'

services:
  # Main API Service
  api:
    image: ${DOCKER_USER:-brainsait}/${IMAGE_NAME:-brainsait-ai}:${VERSION:-latest}
    container_name: brainsait-api
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - GITHUB_TOKEN=${GITHUB_TOKEN}
      - STRIPE_API_KEY=${STRIPE_API_KEY}
      - STRIPE_WEBHOOK_SECRET=${STRIPE_WEBHOOK_SECRET}
      - DATABASE_URL=postgres://brainsait:${DB_PASSWORD}@postgres:5432/brainsait?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - LOG_LEVEL=info
      - ENVIRONMENT=production
    depends_on:
      - postgres
      - redis
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # PostgreSQL Database
  postgres:
    image: postgres:15-alpine
    container_name: brainsait-postgres
    restart: unless-stopped
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_USER=brainsait
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=brainsait
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U brainsait"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis Cache
  redis:
    image: redis:7-alpine
    container_name: brainsait-redis
    restart: unless-stopped
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Prometheus Monitoring
  prometheus:
    image: prom/prometheus:latest
    container_name: brainsait-prometheus
    restart: unless-stopped
    ports:
      - "127.0.0.1:9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'

  # Grafana Dashboards
  grafana:
    image: grafana/grafana:latest
    container_name: brainsait-grafana
    restart: unless-stopped
    ports:
      - "127.0.0.1:3001:3000"
    volumes:
      - grafana_data:/var/lib/grafana
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
      - GF_USERS_ALLOW_SIGN_UP=false
    depends_on:
      - prometheus

volumes:
  postgres_data:
  redis_data:
  prometheus_data:
  grafana_data:
DOCKER_COMPOSE

# Create prometheus config
cat > prometheus.yml << 'PROMETHEUS'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'brainsait-api'
    static_configs:
      - targets: ['api:8080']
    metrics_path: '/metrics'

  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
PROMETHEUS

# Create .env template
echo ""
echo "📝 Step 7: Create Environment Template"
cat > .env.template << 'ENVFILE'
# BrainSait Production Environment
# Copy to .env and fill in values

# GitHub Authentication
GITHUB_TOKEN=your_github_token

# Stripe Billing
STRIPE_API_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...

# Database
DB_PASSWORD=generate_strong_password_here

# Grafana
GRAFANA_PASSWORD=generate_strong_password_here

# Docker Images
DOCKER_USER=brainsait
IMAGE_NAME=brainsait-ai
VERSION=latest
ENVFILE

# Create backup script
echo ""
echo "📝 Step 8: Create Backup Script"
cat > backup.sh << 'BACKUP'
#!/bin/bash
# BrainSait Backup Script

BACKUP_DIR="/opt/brainsait/backups"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# Backup PostgreSQL
docker exec brainsait-postgres pg_dump -U brainsait brainsait | gzip > "${BACKUP_DIR}/db_${DATE}.sql.gz"

# Keep only last 7 days of backups
find $BACKUP_DIR -name "db_*.sql.gz" -mtime +7 -delete

echo "Backup completed: ${BACKUP_DIR}/db_${DATE}.sql.gz"
BACKUP
chmod +x backup.sh

# Setup daily backup cron
(crontab -l 2>/dev/null; echo "0 2 * * * /opt/brainsait/backup.sh") | crontab - 2>/dev/null || true

echo ""
echo "======================"
echo "✅ VM Setup Complete!"
echo ""
echo "Next steps:"
echo ""
echo "1. Create .env file:"
echo "   cp .env.template .env"
echo "   nano .env  # Fill in your values"
echo ""
echo "2. Start services:"
echo "   docker-compose up -d"
echo ""
echo "3. Setup Cloudflare Tunnel:"
echo "   cloudflared tunnel login"
echo "   cloudflared tunnel create brainsait-tunnel"
echo "   cloudflared tunnel route dns brainsait-tunnel api.${DOMAIN}"
echo ""
echo "4. Run Cloudflare Tunnel:"
echo "   cloudflared tunnel run brainsait-tunnel"
echo ""
echo "5. Check status:"
echo "   docker-compose ps"
echo "   curl http://localhost:8080/health"
