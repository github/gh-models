# BrainSait AI Agents

Pre-built AI agents powered by [Docker cagent](https://github.com/docker/cagent) for specialized domains.

## Quick Start

### Install cagent

```bash
# macOS/Linux
brew install docker/tap/cagent

# Or download from releases
curl -L https://github.com/docker/cagent/releases/latest/download/cagent-$(uname -s)-$(uname -m) -o cagent
chmod +x cagent
sudo mv cagent /usr/local/bin/
```

### Set up GitHub Token

```bash
export GITHUB_TOKEN=$(gh auth token)
# Or use your personal access token
export GITHUB_TOKEN=ghp_xxxxx
```

### Run an Agent

```bash
# Arabic Language Assistant
cagent run agents/brainsait-arabic.yaml

# Healthcare Insurance Analyst
cagent run agents/brainsait-healthcare.yaml

# Developer Assistant
cagent run agents/brainsait-developer.yaml
```

## Available Agents

### 🌍 Arabic Language Assistant (`brainsait-arabic.yaml`)

Specialized AI assistant for Arabic speakers:
- **Fluent Arabic** - Modern Standard Arabic and dialects
- **Translation** - Arabic ↔ English translation
- **Cultural awareness** - Respects Arab/Islamic culture
- **Technical help** - Explains concepts in Arabic

**Commands:**
- `/translate` - Translate text to Arabic
- `/explain` - Explain in Arabic  
- `/summarize` - Summarize content in Arabic

### 🏥 Healthcare Insurance Analyst (`brainsait-healthcare.yaml`)

HIPAA-aware assistant for healthcare insurance:
- **Claims Analysis** - Review claims for completeness
- **Prior Authorization** - PA requirements and documentation
- **Medical Coding** - ICD-10, CPT, HCPCS verification
- **Appeals Support** - Help with denial appeals

**Commands:**
- `/analyze-claim` - Analyze insurance claim
- `/check-auth` - Check PA requirements
- `/appeal-help` - Draft appeal letter
- `/code-check` - Verify medical codes

**Multi-agent team:**
- `claims_analyst` - Claims processing specialist
- `prior_auth_specialist` - Prior authorization expert
- `coding_expert` - Medical coding specialist

### 💻 Developer Assistant (`brainsait-developer.yaml`)

AI development team for software engineering:
- **Code Review** - Quality, security, performance
- **Architecture** - System design and decisions
- **Debugging** - Troubleshooting and fixes
- **Testing** - Test generation and strategies
- **DevOps** - CI/CD and deployment

**Commands:**
- `/review` - Review code
- `/explain` - Explain code
- `/refactor` - Suggest improvements
- `/test` - Generate tests
- `/doc` - Generate documentation

**Multi-agent team:**
- `code_reviewer` - Senior code reviewer
- `architect` - Software architect
- `debugger` - Debugging specialist
- `tester` - QA engineer
- `devops` - DevOps engineer

## Integration with BrainSait API

These agents use GitHub Models as the LLM provider, which integrates seamlessly with the BrainSait API:

```yaml
providers:
  github:
    type: openai
    base_url: https://models.inference.ai.azure.com
    api_key: ${GITHUB_TOKEN}
```

To use BrainSait API instead:

```yaml
providers:
  brainsait:
    type: openai
    base_url: https://api.brainsait.ai/v1
    api_key: ${BRAINSAIT_API_KEY}
```

## MCP Tools

Agents use Docker MCP (Model Context Protocol) servers for extended capabilities:

| Tool | Description |
|------|-------------|
| `docker:duckduckgo` | Web search |
| `docker:github-official` | GitHub integration |
| `docker:filesystem` | File operations |

Configure MCP servers in Docker Desktop's MCP Toolkit.

## Creating Custom Agents

1. Copy an existing agent as template
2. Modify the instruction and tools
3. Test with `cagent run your-agent.yaml`
4. Share via Git or Docker Hub

Example minimal agent:

```yaml
agents:
  root:
    model: github-gpt4o
    description: "My Custom Agent"
    instruction: |
      You are a helpful assistant specialized in...

models:
  github-gpt4o:
    provider: github
    model: openai/gpt-4o

providers:
  github:
    type: openai
    base_url: https://models.inference.ai.azure.com
    api_key: ${GITHUB_TOKEN}
```

## Support

- **Documentation**: https://docs.brainsait.ai/agents
- **Issues**: https://github.com/github/gh-models/issues
- **Discord**: https://discord.gg/brainsait
