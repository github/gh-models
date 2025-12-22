#!/bin/bash
# BrainSait Deployment Script
# Builds, pushes to Docker Hub, and deploys to remote VM with Cloudflare Tunnel

set -e

# Configuration
DOCKER_REGISTRY="${DOCKER_REGISTRY:-docker.io}"
DOCKER_USERNAME="${DOCKER_USERNAME:-brainsait}"
IMAGE_NAME="${IMAGE_NAME:-brainsait-ai}"
VERSION="${VERSION:-latest}"
REMOTE_HOST="${REMOTE_HOST:-}"
REMOTE_USER="${REMOTE_USER:-root}"
CF_TUNNEL_TOKEN="${CF_TUNNEL_TOKEN:-}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check required environment variables
check_requirements() {
    log_info "Checking requirements..."
    
    if [ -z "$DOCKER_USERNAME" ]; then
        log_error "DOCKER_USERNAME is required"
        exit 1
    fi
    
    if [ -z "$REMOTE_HOST" ]; then
        log_error "REMOTE_HOST is required for deployment"
        exit 1
    fi
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    log_info "All requirements met"
}

# Build Docker image
build_image() {
    log_info "Building Docker image..."
    
    docker build \
        --platform linux/amd64 \
        -t "$DOCKER_REGISTRY/$DOCKER_USERNAME/$IMAGE_NAME:$VERSION" \
        -t "$DOCKER_REGISTRY/$DOCKER_USERNAME/$IMAGE_NAME:latest" \
        -f Dockerfile \
        .
    
    log_info "Docker image built successfully"
}

# Push to Docker Hub
push_image() {
    log_info "Pushing image to Docker Hub..."
    
    # Login to Docker Hub
    if [ -n "$DOCKER_PASSWORD" ]; then
        echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
    else
        log_warn "DOCKER_PASSWORD not set, assuming already logged in"
    fi
    
    docker push "$DOCKER_REGISTRY/$DOCKER_USERNAME/$IMAGE_NAME:$VERSION"
    docker push "$DOCKER_REGISTRY/$DOCKER_USERNAME/$IMAGE_NAME:latest"
    
    log_info "Image pushed successfully"
}

# Deploy to remote VM
deploy_remote() {
    log_info "Deploying to remote VM ($REMOTE_HOST)..."
    
    # Create deployment script
    DEPLOY_SCRIPT=$(cat <<'EOF'
#!/bin/bash
set -e

# Pull latest images
docker-compose pull

# Stop existing containers
docker-compose down

# Start new containers
docker-compose up -d

# Clean up old images
docker image prune -f

# Check health
sleep 10
curl -f http://localhost:8080/health || exit 1

echo "Deployment successful!"
EOF
)
    
    # Copy docker-compose and .env files
    scp docker-compose.yml "$REMOTE_USER@$REMOTE_HOST:~/brainsait/"
    scp .env.production "$REMOTE_USER@$REMOTE_HOST:~/brainsait/.env"
    
    # Execute deployment
    ssh "$REMOTE_USER@$REMOTE_HOST" "cd ~/brainsait && $DEPLOY_SCRIPT"
    
    log_info "Remote deployment completed"
}

# Setup Cloudflare Tunnel
setup_cloudflare_tunnel() {
    log_info "Setting up Cloudflare Tunnel..."
    
    if [ -z "$CF_TUNNEL_TOKEN" ]; then
        log_warn "CF_TUNNEL_TOKEN not set, skipping tunnel setup"
        return
    fi
    
    # Create cloudflared config on remote
    CF_CONFIG=$(cat <<EOF
tunnel: brainsait
credentials-file: /etc/cloudflared/credentials.json

ingress:
  - hostname: api.brainsait.ai
    service: http://localhost:8080
  - hostname: docs.brainsait.ai
    service: http://localhost:8081
  - service: http_status:404
EOF
)
    
    ssh "$REMOTE_USER@$REMOTE_HOST" "mkdir -p /etc/cloudflared && echo '$CF_CONFIG' > /etc/cloudflared/config.yml"
    
    # Install and start cloudflared
    ssh "$REMOTE_USER@$REMOTE_HOST" << 'REMOTE_EOF'
# Install cloudflared if not present
if ! command -v cloudflared &> /dev/null; then
    curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
    sudo dpkg -i cloudflared.deb
    rm cloudflared.deb
fi

# Setup tunnel as service
sudo cloudflared service install $CF_TUNNEL_TOKEN

# Start tunnel
sudo systemctl start cloudflared
sudo systemctl enable cloudflared
REMOTE_EOF
    
    log_info "Cloudflare Tunnel configured"
}

# Print usage
usage() {
    cat << EOF
BrainSait Deployment Script

Usage: $0 [command]

Commands:
    build       Build Docker image only
    push        Push image to Docker Hub
    deploy      Deploy to remote VM
    tunnel      Setup Cloudflare Tunnel
    all         Run all steps (default)
    
Environment Variables:
    DOCKER_USERNAME     Docker Hub username (required)
    DOCKER_PASSWORD     Docker Hub password (for push)
    IMAGE_NAME          Docker image name (default: brainsait-ai)
    VERSION             Image version tag (default: latest)
    REMOTE_HOST         Remote VM hostname/IP (required for deploy)
    REMOTE_USER         Remote VM user (default: root)
    CF_TUNNEL_TOKEN     Cloudflare Tunnel token

Examples:
    # Build and push
    DOCKER_USERNAME=myuser VERSION=v1.0.0 $0 build push
    
    # Full deployment
    DOCKER_USERNAME=myuser REMOTE_HOST=server.example.com $0 all
EOF
}

# Main
main() {
    if [ $# -eq 0 ]; then
        set -- "all"
    fi
    
    for cmd in "$@"; do
        case "$cmd" in
            build)
                build_image
                ;;
            push)
                push_image
                ;;
            deploy)
                check_requirements
                deploy_remote
                ;;
            tunnel)
                setup_cloudflare_tunnel
                ;;
            all)
                check_requirements
                build_image
                push_image
                deploy_remote
                setup_cloudflare_tunnel
                ;;
            -h|--help|help)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown command: $cmd"
                usage
                exit 1
                ;;
        esac
    done
    
    log_info "All tasks completed successfully!"
}

main "$@"
