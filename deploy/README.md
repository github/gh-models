# BrainSait Deployment Guide

This guide covers the complete deployment workflow for the BrainSait AI Platform.

## Overview

The deployment process follows these steps:

1. **Build** - Create Docker images locally
2. **Push** - Upload images to Docker Hub
3. **Deploy** - Pull and run images on a remote VM
4. **Expose** - Use Cloudflare Tunnel to expose services

## Prerequisites

- Docker and Docker Compose installed locally
- Docker Hub account
- Remote VM (Ubuntu 22.04 LTS recommended)
- Cloudflare account with a domain
- GitHub Personal Access Token (for GitHub Models API)
- Stripe account (for billing)

## Quick Start

### 1. Build Images Locally

```bash
# Build the main CLI image
docker build -t brainsait/brainsait-ai:latest .

# Build the API server image
docker build -f Dockerfile.api -t brainsait/brainsait-ai-api:latest .
```

### 2. Push to Docker Hub

```bash
# Login to Docker Hub
docker login

# Push images
./deploy/docker-hub-push.sh
```

Or manually:

```bash
docker push brainsait/brainsait-ai:latest
docker push brainsait/brainsait-ai-api:latest
```

### 3. Setup Remote VM

SSH into your VM and run:

```bash
# Download and run setup script
curl -fsSL https://raw.githubusercontent.com/your-org/brainsait/main/deploy/vm-setup.sh | bash
```

Or manually:

```bash
# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo apt install -y docker-compose-plugin

# Install Cloudflared
curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared.deb
```

### 4. Configure Environment

```bash
cd /opt/brainsait
cp .env.template .env
nano .env
```

Required environment variables:

| Variable | Description |
|----------|-------------|
| `GITHUB_TOKEN` | GitHub PAT with models access |
| `STRIPE_API_KEY` | Stripe secret key |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook signing secret |
| `DB_PASSWORD` | PostgreSQL password |
| `GRAFANA_PASSWORD` | Grafana admin password |

### 5. Start Services

```bash
docker-compose up -d

# Check status
docker-compose ps
docker-compose logs -f api
```

### 6. Setup Cloudflare Tunnel

```bash
# Login to Cloudflare
cloudflared tunnel login

# Create tunnel
cloudflared tunnel create brainsait-tunnel

# Configure DNS routes
cloudflared tunnel route dns brainsait-tunnel api.yourdomain.com
cloudflared tunnel route dns brainsait-tunnel docs.yourdomain.com

# Install as service
sudo cloudflared service install

# Start tunnel
sudo systemctl start cloudflared
sudo systemctl enable cloudflared
```

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                      Cloudflare Edge                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ api.domain  │  │docs.domain  │  │ grafana.domain      │  │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘  │
└─────────┼────────────────┼────────────────────┼──────────────┘
          │                │                    │
          ▼                ▼                    ▼
┌─────────────────────────────────────────────────────────────┐
│                   Cloudflare Tunnel                          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      Remote VM                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                  Docker Network                      │    │
│  │  ┌─────────┐  ┌──────────┐  ┌───────┐  ┌─────────┐  │    │
│  │  │   API   │──│ Postgres │  │ Redis │  │Prometheus│  │    │
│  │  │  :8080  │  │  :5432   │  │ :6379 │  │  :9090  │  │    │
│  │  └─────────┘  └──────────┘  └───────┘  └─────────┘  │    │
│  │                                        ┌─────────┐  │    │
│  │                                        │ Grafana │  │    │
│  │                                        │  :3001  │  │    │
│  │                                        └─────────┘  │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## API Endpoints

Once deployed, the following endpoints are available:

### Public Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/v1/chat` | POST | Chat completion |
| `/v1/chat/stream` | POST | Streaming chat |
| `/v1/models` | GET | List available models |
| `/v1/models/{id}` | GET | Get model details |
| `/v1/prompts` | GET | List prompt templates |
| `/v1/prompts/{domain}` | GET | Domain-specific prompts |

### Authenticated Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/user` | GET | Get user profile |
| `/v1/user/usage` | GET | Get usage statistics |
| `/v1/billing/checkout` | POST | Create Stripe checkout |
| `/v1/billing/portal` | POST | Create billing portal |

## Monitoring

### Grafana Dashboards

Access Grafana at `https://grafana.yourdomain.com` with:
- Username: `admin`
- Password: (from `GRAFANA_PASSWORD` in .env)

### Prometheus Metrics

The API server exposes metrics at `/metrics`:
- Request counts and latency
- Token usage
- Error rates

## Backup & Recovery

### Automated Backups

The VM setup configures daily PostgreSQL backups at 2 AM:

```bash
# Manual backup
/opt/brainsait/backup.sh

# View backups
ls -la /opt/brainsait/backups/
```

### Restore from Backup

```bash
# Stop services
docker-compose down

# Restore database
gunzip < backups/db_YYYYMMDD_HHMMSS.sql.gz | \
  docker exec -i brainsait-postgres psql -U brainsait brainsait

# Start services
docker-compose up -d
```

## Troubleshooting

### Check Service Status

```bash
# Container status
docker-compose ps

# Container logs
docker-compose logs -f api
docker-compose logs -f postgres

# Tunnel status
sudo systemctl status cloudflared
```

### Common Issues

**API not responding:**
```bash
# Check health
curl http://localhost:8080/health

# Check logs
docker-compose logs api
```

**Database connection failed:**
```bash
# Check PostgreSQL
docker-compose logs postgres
docker exec -it brainsait-postgres psql -U brainsait -d brainsait
```

**Tunnel not working:**
```bash
# Check tunnel status
cloudflared tunnel info brainsait-tunnel

# Check tunnel logs
sudo journalctl -u cloudflared -f
```

## Scaling

### Horizontal Scaling

To run multiple API instances:

```yaml
# In docker-compose.yml
services:
  api:
    deploy:
      replicas: 3
```

### Load Balancing

Cloudflare automatically load balances across tunnel instances.

## Security Checklist

- [ ] Use strong passwords in `.env`
- [ ] Enable Cloudflare Access for admin endpoints
- [ ] Configure firewall (only allow ports 22, 80, 443)
- [ ] Enable automatic security updates
- [ ] Set up SSL certificate rotation
- [ ] Enable audit logging
- [ ] Configure rate limiting in Cloudflare

## Support

For issues or questions:
- GitHub Issues: https://github.com/your-org/brainsait/issues
- Documentation: https://docs.brainsait.com
- Email: support@brainsait.com
