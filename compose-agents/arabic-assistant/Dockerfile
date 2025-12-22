# BrainSait Arabic Assistant Agent
FROM python:3.12-slim

WORKDIR /app

# Install dependencies
RUN pip install --no-cache-dir \
    google-adk \
    httpx \
    uvicorn

# Copy agent code
COPY agents/ ./agents/

# Set environment
ENV PYTHONUNBUFFERED=1
ENV LANGUAGE=ar

# Run the agent
CMD ["python", "-m", "google.adk.cli", "web", "--agents-dir", "agents", "--host", "0.0.0.0", "--port", "8080"]
