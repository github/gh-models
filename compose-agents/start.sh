#!/bin/bash
# BrainSait Compose Agents - Quick Start Script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC}          ${GREEN}BrainSait Multi-Agent AI Platform${NC}               ${BLUE}║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_menu() {
    echo -e "${YELLOW}Available Agent Systems:${NC}"
    echo ""
    echo "  1) 🌍 Arabic Assistant      - Translation, writing, research in Arabic"
    echo "  2) 🏥 Healthcare Insurance  - Claims analysis, PA, coding, appeals"
    echo "  3) 💻 Developer Assistant   - Code review, architecture, debugging"
    echo "  4) 🔍 SQL Analyst           - Natural language to SQL queries"
    echo ""
    echo "  q) Quit"
    echo ""
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}Error: Docker is not installed.${NC}"
        echo "Please install Docker Desktop: https://www.docker.com/products/docker-desktop/"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        echo -e "${RED}Error: Docker daemon is not running.${NC}"
        echo "Please start Docker Desktop."
        exit 1
    fi
}

check_github_token() {
    if [ -z "$GITHUB_TOKEN" ]; then
        echo -e "${YELLOW}GitHub token not set.${NC}"
        
        # Try to get from gh CLI
        if command -v gh &> /dev/null; then
            export GITHUB_TOKEN=$(gh auth token 2>/dev/null)
            if [ -n "$GITHUB_TOKEN" ]; then
                echo -e "${GREEN}✓ Using GitHub token from gh CLI${NC}"
                return 0
            fi
        fi
        
        echo ""
        echo "To use GitHub Models, set your token:"
        echo "  export GITHUB_TOKEN=\$(gh auth token)"
        echo "  # or"
        echo "  export GITHUB_TOKEN=ghp_xxxxx"
        echo ""
        return 1
    fi
    echo -e "${GREEN}✓ GitHub token configured${NC}"
    return 0
}

run_agent() {
    local agent_dir=$1
    local agent_name=$2
    local use_github=${3:-false}
    
    echo ""
    echo -e "${GREEN}Starting $agent_name...${NC}"
    echo ""
    
    cd "$SCRIPT_DIR/$agent_dir"
    
    if [ "$use_github" = true ] && [ -f "compose.github.yaml" ]; then
        echo "Using GitHub Models..."
        docker compose -f compose.yaml -f compose.github.yaml up --build
    else
        echo "Using Docker Model Runner (local)..."
        docker compose up --build
    fi
}

main() {
    print_header
    check_docker
    
    local has_token=false
    if check_github_token; then
        has_token=true
    fi
    
    echo ""
    
    while true; do
        print_menu
        read -p "Select an agent (1-4, or q to quit): " choice
        
        case $choice in
            1)
                run_agent "arabic-assistant" "Arabic Language Assistant" $has_token
                ;;
            2)
                run_agent "healthcare-insurance" "Healthcare Insurance Analyst" $has_token
                ;;
            3)
                run_agent "developer-assistant" "Developer Assistant" $has_token
                ;;
            4)
                run_agent "sql-analyst" "SQL Analyst" $has_token
                ;;
            q|Q)
                echo ""
                echo -e "${GREEN}Goodbye!${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}Invalid selection. Please try again.${NC}"
                echo ""
                ;;
        esac
    done
}

main "$@"
