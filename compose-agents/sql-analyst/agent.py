"""BrainSait SQL Analyst - LangGraph Agent for Natural Language to SQL"""

import os
import json
from typing import TypedDict, Annotated, Sequence
from langgraph.graph import StateGraph, END
from langchain_core.messages import HumanMessage, AIMessage, BaseMessage
import httpx
import uvicorn
from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse, JSONResponse
from fastapi.staticfiles import StaticFiles

# State definition
class AgentState(TypedDict):
    messages: Annotated[Sequence[BaseMessage], lambda x, y: x + y]
    question: str
    sql_query: str
    query_result: str
    final_answer: str


# LLM client
class LLMClient:
    def __init__(self):
        self.base_url = os.environ.get("MODEL_RUNNER_URL", "http://localhost:8080/v1")
        self.model = os.environ.get("MODEL_RUNNER_MODEL", "ai/qwen3:14B-Q6_K")
    
    async def chat(self, messages: list[dict], system_prompt: str = None) -> str:
        async with httpx.AsyncClient() as client:
            payload = {
                "model": self.model,
                "messages": messages,
                "temperature": 0.1,
            }
            if system_prompt:
                payload["messages"] = [{"role": "system", "content": system_prompt}] + payload["messages"]
            
            response = await client.post(
                f"{self.base_url}/chat/completions",
                json=payload,
                timeout=120.0
            )
            response.raise_for_status()
            return response.json()["choices"][0]["message"]["content"]


# MCP client for database queries
class MCPClient:
    def __init__(self):
        self.mcp_url = os.environ.get("MCP_SERVER_URL", "http://mcp-gateway:8811/sse")
    
    async def query(self, sql: str) -> str:
        async with httpx.AsyncClient() as client:
            response = await client.post(
                self.mcp_url.replace("/sse", "/tools/query"),
                json={"sql": sql},
                timeout=60.0
            )
            if response.status_code == 200:
                return json.dumps(response.json(), indent=2)
            return f"Error: {response.text}"


llm = LLMClient()
mcp = MCPClient()


# Agent nodes
async def understand_question(state: AgentState) -> dict:
    """Understand the user's question and identify required data."""
    system_prompt = """You are a SQL expert. Analyze the user's question and identify:
1. What data they're looking for
2. What tables might be relevant
3. Any filters or aggregations needed

Respond with a brief analysis."""

    messages = [{"role": "user", "content": state["question"]}]
    response = await llm.chat(messages, system_prompt)
    
    return {
        "messages": [AIMessage(content=f"Analysis: {response}")]
    }


async def generate_sql(state: AgentState) -> dict:
    """Generate SQL query based on the question."""
    system_prompt = """You are a SQL expert for PostgreSQL. Generate a SQL query to answer the user's question.

Rules:
- Use only SELECT statements (no INSERT, UPDATE, DELETE)
- Include appropriate JOINs if needed
- Add LIMIT 100 to prevent large result sets
- Format the query nicely

Respond with ONLY the SQL query, no explanation."""

    messages = [{"role": "user", "content": state["question"]}]
    sql_query = await llm.chat(messages, system_prompt)
    
    # Clean up the query
    sql_query = sql_query.strip()
    if sql_query.startswith("```"):
        sql_query = sql_query.split("```")[1]
        if sql_query.startswith("sql"):
            sql_query = sql_query[3:]
    sql_query = sql_query.strip()
    
    return {
        "sql_query": sql_query,
        "messages": [AIMessage(content=f"Generated SQL:\n```sql\n{sql_query}\n```")]
    }


async def execute_query(state: AgentState) -> dict:
    """Execute the SQL query against the database."""
    result = await mcp.query(state["sql_query"])
    
    return {
        "query_result": result,
        "messages": [AIMessage(content=f"Query Result:\n{result}")]
    }


async def generate_answer(state: AgentState) -> dict:
    """Generate a natural language answer from the query results."""
    system_prompt = """You are a data analyst. Given the user's question and the query results, 
provide a clear, natural language answer. Include relevant numbers and insights."""

    messages = [
        {"role": "user", "content": f"Question: {state['question']}\n\nQuery Results:\n{state['query_result']}"}
    ]
    answer = await llm.chat(messages, system_prompt)
    
    return {
        "final_answer": answer,
        "messages": [AIMessage(content=answer)]
    }


# Build the graph
def build_graph():
    workflow = StateGraph(AgentState)
    
    workflow.add_node("understand", understand_question)
    workflow.add_node("generate_sql", generate_sql)
    workflow.add_node("execute", execute_query)
    workflow.add_node("answer", generate_answer)
    
    workflow.set_entry_point("understand")
    workflow.add_edge("understand", "generate_sql")
    workflow.add_edge("generate_sql", "execute")
    workflow.add_edge("execute", "answer")
    workflow.add_edge("answer", END)
    
    return workflow.compile()


# FastAPI app
app = FastAPI(title="BrainSait SQL Analyst")
graph = build_graph()


@app.get("/", response_class=HTMLResponse)
async def index():
    return """
    <!DOCTYPE html>
    <html>
    <head>
        <title>BrainSait SQL Analyst</title>
        <style>
            body { font-family: system-ui; max-width: 800px; margin: 0 auto; padding: 20px; }
            .chat { border: 1px solid #ddd; border-radius: 8px; padding: 20px; margin: 20px 0; }
            .message { margin: 10px 0; padding: 10px; border-radius: 4px; }
            .user { background: #e3f2fd; }
            .assistant { background: #f5f5f5; }
            input[type="text"] { width: 100%; padding: 10px; margin: 10px 0; }
            button { padding: 10px 20px; background: #1976d2; color: white; border: none; cursor: pointer; }
            pre { background: #f5f5f5; padding: 10px; overflow-x: auto; }
        </style>
    </head>
    <body>
        <h1>🔍 BrainSait SQL Analyst</h1>
        <p>Ask questions about your data in natural language.</p>
        
        <form id="queryForm">
            <input type="text" id="question" placeholder="e.g., What were the top 5 products by sales last month?" />
            <button type="submit">Ask</button>
        </form>
        
        <div id="chat" class="chat"></div>
        
        <script>
            document.getElementById('queryForm').addEventListener('submit', async (e) => {
                e.preventDefault();
                const question = document.getElementById('question').value;
                const chat = document.getElementById('chat');
                
                chat.innerHTML += `<div class="message user"><strong>You:</strong> ${question}</div>`;
                chat.innerHTML += `<div class="message assistant"><em>Analyzing...</em></div>`;
                
                const response = await fetch('/query', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({question})
                });
                
                const result = await response.json();
                chat.lastElementChild.innerHTML = `
                    <strong>SQL Analyst:</strong>
                    <pre>${result.sql_query}</pre>
                    <p>${result.answer}</p>
                `;
                
                document.getElementById('question').value = '';
            });
        </script>
    </body>
    </html>
    """


@app.post("/query")
async def query(request: Request):
    data = await request.json()
    question = data.get("question", "")
    
    initial_state = {
        "messages": [HumanMessage(content=question)],
        "question": question,
        "sql_query": "",
        "query_result": "",
        "final_answer": ""
    }
    
    result = await graph.ainvoke(initial_state)
    
    return JSONResponse({
        "question": question,
        "sql_query": result["sql_query"],
        "result": result["query_result"],
        "answer": result["final_answer"]
    })


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8080)
