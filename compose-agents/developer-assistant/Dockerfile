# BrainSait Developer Assistant
FROM python:3.12-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
RUN pip install --no-cache-dir \
    google-adk \
    httpx \
    uvicorn \
    gitpython

# Copy agent code
COPY agents/ ./agents/

# Set environment
ENV PYTHONUNBUFFERED=1

# Run the agent
CMD ["python", "-m", "google.adk.cli", "web", "--agents-dir", "agents", "--host", "0.0.0.0", "--port", "8080"]
