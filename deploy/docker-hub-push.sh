#!/bin/bash
# Docker Hub Push Script for BrainSait
# Builds multi-architecture images and pushes to Docker Hub

set -e

# Configuration
DOCKER_USER="${DOCKER_USER:-brainsait}"
IMAGE_NAME="${IMAGE_NAME:-brainsait-ai}"
VERSION="${VERSION:-$(git describe --tags --always 2>/dev/null || echo 'latest')}"
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"

echo "🐳 BrainSait Docker Build & Push"
echo "=================================="
echo "User: $DOCKER_USER"
echo "Image: $IMAGE_NAME"
echo "Version: $VERSION"
echo "Platforms: $PLATFORMS"
echo ""

# Check if logged into Docker Hub
if ! docker info 2>/dev/null | grep -q "Username"; then
    echo "📝 Please login to Docker Hub:"
    docker login
fi

# Setup buildx for multi-platform builds
echo ""
echo "🔧 Setting up Docker Buildx..."
if ! docker buildx inspect brainsait-builder &>/dev/null; then
    docker buildx create --name brainsait-builder --driver docker-container --bootstrap
fi
docker buildx use brainsait-builder

# Build and push main application image
echo ""
echo "🏗️  Building and pushing ${DOCKER_USER}/${IMAGE_NAME}:${VERSION}..."

cd "$(dirname "$0")/.."

docker buildx build \
    --platform "$PLATFORMS" \
    --tag "${DOCKER_USER}/${IMAGE_NAME}:${VERSION}" \
    --tag "${DOCKER_USER}/${IMAGE_NAME}:latest" \
    --push \
    --file Dockerfile \
    .

echo "✅ Pushed: ${DOCKER_USER}/${IMAGE_NAME}:${VERSION}"
echo "✅ Pushed: ${DOCKER_USER}/${IMAGE_NAME}:latest"

# Build and push API server image (if separate Dockerfile exists)
if [ -f "Dockerfile.api" ]; then
    echo ""
    echo "🏗️  Building and pushing ${DOCKER_USER}/${IMAGE_NAME}-api:${VERSION}..."
    
    docker buildx build \
        --platform "$PLATFORMS" \
        --tag "${DOCKER_USER}/${IMAGE_NAME}-api:${VERSION}" \
        --tag "${DOCKER_USER}/${IMAGE_NAME}-api:latest" \
        --push \
        --file Dockerfile.api \
        .
    
    echo "✅ Pushed: ${DOCKER_USER}/${IMAGE_NAME}-api:${VERSION}"
fi

echo ""
echo "=================================="
echo "✅ Docker Hub Push Complete!"
echo ""
echo "Pull commands:"
echo "  docker pull ${DOCKER_USER}/${IMAGE_NAME}:${VERSION}"
echo "  docker pull ${DOCKER_USER}/${IMAGE_NAME}:latest"
echo ""
echo "Run command:"
echo "  docker run -p 8080:8080 ${DOCKER_USER}/${IMAGE_NAME}:${VERSION}"
