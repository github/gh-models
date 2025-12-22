#!/bin/bash
# Cloudflare Tunnel Setup Script for BrainSait
# Prerequisites: cloudflared CLI installed and authenticated

set -e

TUNNEL_NAME="${TUNNEL_NAME:-brainsait-tunnel}"
DOMAIN="${DOMAIN:-brainsait.com}"

echo "🌐 BrainSait Cloudflare Tunnel Setup"
echo "======================================"

# Check if cloudflared is installed
if ! command -v cloudflared &> /dev/null; then
    echo "❌ cloudflared CLI not found. Installing..."
    
    # Detect OS and install
    if [[ "$OSTYPE" == "darwin"* ]]; then
        brew install cloudflare/cloudflare/cloudflared
    elif [[ -f /etc/debian_version ]]; then
        curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
        sudo dpkg -i cloudflared.deb
        rm cloudflared.deb
    elif [[ -f /etc/redhat-release ]]; then
        curl -L --output cloudflared.rpm https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-x86_64.rpm
        sudo rpm -i cloudflared.rpm
        rm cloudflared.rpm
    else
        echo "❌ Unsupported OS. Please install cloudflared manually."
        exit 1
    fi
fi

echo "✅ cloudflared installed: $(cloudflared --version)"

# Login to Cloudflare (if not already)
echo ""
echo "📝 Step 1: Authenticate with Cloudflare"
if [ ! -f ~/.cloudflared/cert.pem ]; then
    echo "Opening browser for authentication..."
    cloudflared tunnel login
else
    echo "Already authenticated"
fi

# Create tunnel
echo ""
echo "📝 Step 2: Create tunnel '$TUNNEL_NAME'"
if cloudflared tunnel list | grep -q "$TUNNEL_NAME"; then
    echo "Tunnel '$TUNNEL_NAME' already exists"
    TUNNEL_ID=$(cloudflared tunnel list | grep "$TUNNEL_NAME" | awk '{print $1}')
else
    cloudflared tunnel create "$TUNNEL_NAME"
    TUNNEL_ID=$(cloudflared tunnel list | grep "$TUNNEL_NAME" | awk '{print $1}')
    echo "Created tunnel with ID: $TUNNEL_ID"
fi

# Configure DNS routes
echo ""
echo "📝 Step 3: Configure DNS routes"

configure_dns() {
    local subdomain=$1
    local hostname="${subdomain}.${DOMAIN}"
    
    if cloudflared tunnel route dns "$TUNNEL_NAME" "$hostname" 2>/dev/null; then
        echo "✅ Configured: $hostname"
    else
        echo "⚠️  Route may already exist for: $hostname"
    fi
}

configure_dns "api"
configure_dns "docs"
configure_dns "metrics"
configure_dns "grafana"

# Create credentials directory
echo ""
echo "📝 Step 4: Setup credentials"
sudo mkdir -p /etc/cloudflared
sudo cp ~/.cloudflared/${TUNNEL_ID}.json /etc/cloudflared/credentials.json
sudo chmod 600 /etc/cloudflared/credentials.json

# Copy configuration
echo ""
echo "📝 Step 5: Install configuration"
sudo cp "$(dirname "$0")/config.yml" /etc/cloudflared/config.yml
sudo sed -i "s/brainsait-tunnel/$TUNNEL_ID/g" /etc/cloudflared/config.yml 2>/dev/null || \
sudo sed -i '' "s/brainsait-tunnel/$TUNNEL_ID/g" /etc/cloudflared/config.yml

# Install as service
echo ""
echo "📝 Step 6: Install as system service"
sudo cloudflared service install

echo ""
echo "======================================"
echo "✅ Cloudflare Tunnel Setup Complete!"
echo ""
echo "Tunnel ID: $TUNNEL_ID"
echo "Endpoints configured:"
echo "  - https://api.${DOMAIN} -> localhost:8080 (API)"
echo "  - https://docs.${DOMAIN} -> localhost:8081 (Docs)"
echo "  - https://metrics.${DOMAIN} -> localhost:3000 (Metrics)"
echo "  - https://grafana.${DOMAIN} -> localhost:3001 (Grafana)"
echo ""
echo "Commands:"
echo "  Start:   sudo systemctl start cloudflared"
echo "  Stop:    sudo systemctl stop cloudflared"
echo "  Status:  sudo systemctl status cloudflared"
echo "  Logs:    sudo journalctl -u cloudflared -f"
echo ""
echo "Manual run: cloudflared tunnel run $TUNNEL_NAME"
