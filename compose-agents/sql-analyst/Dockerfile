# BrainSait SQL Analyst
FROM python:3.12-slim

WORKDIR /app

RUN pip install --no-cache-dir \
    langgraph \
    langchain-core \
    httpx \
    uvicorn \
    psycopg2-binary

COPY agent.py ./

ENV PYTHONUNBUFFERED=1

CMD ["python", "agent.py"]
