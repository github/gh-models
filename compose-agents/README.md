# BrainSait Compose Agents

Multi-agent AI systems built with [Docker Compose for Agents](https://github.com/docker/compose-for-agents).

## Overview

BrainSait provides pre-built, domain-specific AI agent systems that run locally using Docker Model Runner or in the cloud with OpenAI-compatible APIs.

## Available Agent Systems

| Agent System | Description | Framework | Models |
|--------------|-------------|-----------|--------|
| [Arabic Assistant](./arabic-assistant/) | Arabic language translation, writing, and research | Google ADK | qwen3 |
| [Healthcare Insurance](./healthcare-insurance/) | Claims analysis, PA, coding, appeals | Google ADK | qwen3 |
| [Developer Assistant](./developer-assistant/) | Code review, architecture, debugging, testing, DevOps | Google ADK | qwen3 |
| [SQL Analyst](./sql-analyst/) | Natural language to SQL with data analysis | LangGraph | qwen3 |

## Prerequisites

- **Docker Desktop 4.43.0+** with Model Runner enabled
- **GPU recommended** for running local models (or use [Docker Offload](https://www.docker.com/products/docker-offload/))
- **MCP Toolkit** enabled in Docker Desktop settings

## Quick Start

### 1. Choose an Agent System

```bash
cd compose-agents/developer-assistant
```

### 2. Configure Secrets (if required)

```bash
# Copy example env file
cp mcp.env.example .mcp.env

# Edit with your tokens
nano .mcp.env
```

### 3. Start the Agent System

```bash
docker compose up --build
```

### 4. Access the Web Interface

Open [http://localhost:8080](http://localhost:8080) in your browser.

## Using OpenAI Instead of Local Models

To use OpenAI models instead of running locally:

1. Create a secret file:

```bash
echo "sk-your-openai-key" > secret.openai-api-key
```

2. Start with OpenAI configuration:

```bash
docker compose -f compose.yaml -f compose.openai.yaml up
```

## Using GitHub Models

To use GitHub Models as the LLM provider:

```bash
# Set your GitHub token
export GITHUB_TOKEN=$(gh auth token)

# Start with GitHub Models
docker compose -f compose.yaml -f compose.github.yaml up
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    User Interface (Port 8080)                │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Agent Framework                         │
│              (Google ADK / LangGraph / CrewAI)              │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
       ┌──────────┐    ┌──────────┐    ┌──────────┐
       │  Agent 1 │    │  Agent 2 │    │  Agent N │
       └──────────┘    └──────────┘    └──────────┘
              │               │               │
              └───────────────┼───────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                       MCP Gateway                            │
│              (Tool orchestration & security)                 │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
       ┌──────────┐    ┌──────────┐    ┌──────────┐
       │ DuckDuck │    │  GitHub  │    │ Postgres │
       │   Go     │    │ Official │    │   MCP    │
       └──────────┘    └──────────┘    └──────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Docker Model Runner                       │
│                (Local LLM inference engine)                  │
└─────────────────────────────────────────────────────────────┘
```

## Creating Custom Agent Systems

### 1. Create Directory Structure

```bash
mkdir -p my-agent/agents
```

### 2. Create Agent Definition

```python
# my-agent/agents/agent.py
from google.adk.agents import Agent

root_agent = Agent(
    name="my_agent",
    description="My custom agent",
    instruction="""You are a helpful assistant...""",
    model="qwen3",
)
```

### 3. Create Compose File

```yaml
# my-agent/compose.yaml
services:
  agent:
    build: .
    ports:
      - 8080:8080
    environment:
      - MCP_SERVER_URL=http://mcp-gateway:8811/sse
    depends_on:
      - mcp-gateway
    models:
      qwen3:
        endpoint_var: MODEL_RUNNER_URL
        model_var: MODEL_RUNNER_MODEL

  mcp-gateway:
    image: docker/mcp-gateway:latest
    use_api_socket: true
    command:
      - --transport=sse
      - --servers=duckduckgo

models:
  qwen3:
    model: ai/qwen3:14B-Q6_K
```

### 4. Create Dockerfile

```dockerfile
FROM python:3.12-slim
WORKDIR /app
RUN pip install google-adk httpx uvicorn
COPY agents/ ./agents/
CMD ["python", "-m", "google.adk.cli", "web", "--agents-dir", "agents", "--host", "0.0.0.0", "--port", "8080"]
```

### 5. Run Your Agent

```bash
cd my-agent
docker compose up --build
```

## MCP Tools Reference

| Server | Tools | Description |
|--------|-------|-------------|
| duckduckgo | search, fetch_content | Web search |
| github-official | read_file, list_repos, search_code | GitHub integration |
| filesystem | read_file, write_file, list_directory | File operations |
| postgres | query | PostgreSQL database |
| brave | search | Brave search engine |
| wikipedia-mcp | search, get_article | Wikipedia access |

## Integration with BrainSait CLI

These agent systems complement the `gh models` CLI:

```bash
# Generate tests for your prompts
gh models generate my-prompt.yml

# Evaluate prompt performance
gh models eval my-prompt.yml

# Run inference
gh models run my-prompt.yml
```

## Troubleshooting

### Model Runner Not Starting

Ensure Docker Desktop has Model Runner enabled:

1. Open Docker Desktop settings
2. Go to "Features in development"
3. Enable "Docker Model Runner"

### Out of Memory

Reduce context size in compose.yaml:

```yaml
models:
  qwen3:
    model: ai/qwen3:4B-Q4_0  # Smaller model
    context_size: 4096       # Smaller context
```

### MCP Gateway Connection Failed

Check that MCP Toolkit is enabled in Docker Desktop:

1. Settings → Features in development → Enable MCP Toolkit
2. Restart Docker Desktop

## License

Apache 2.0 OR MIT (dual-licensed)
