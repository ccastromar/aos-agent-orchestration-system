# AOS Banking v2 (with LLM / Ollama + Qwen3)

Domain: retail banking (demo)

* Check balance
* Check transactions
* Send Bizum

This version adds:

* **LLM Intent Engine** (Ollama + `qwen3`) to understand natural language:

  * "What’s my balance?"
  * "Show me the January transactions"
  * "Send a 25€ Bizum to Laura"
* **Analyst LLM**: generates a human summary of the operation and its risk.
* **Deterministic YAML pipelines**: without the LLM touching the flow or the tools.
* **HTTP Tools** that call a simulated banking backend.

## Architecture

* `internal/bus`: in-memory bus between agents.
* `internal/llm`: Ollama client (`qwen3`) + intent & analyst functions.
* `internal/config`: loads YAML from `config/tools`, `config/pipelines`, `config/intents`.
* `internal/agent`:

  * `api_agent`: exposes `/ask` (operation+params) and `/ask_nlp` (natural message).
  * `inspector`: creates the base task and sends it to the Planner.
  * `planner`: if an `operation` is provided, maps it directly; if a `message` is provided, uses LLM-Intent.
  * `verifier`: executes pipeline steps using declared tools.
  * `analyst`: calls the LLM to generate a summary and stores the final result.

## Run demo

### 1. Simulated banking backend

```bash
cd cmd/mock-backend
go run .
# listens on :9000
```

### 2. AOS-Banking

```bash
cd cmd/aos-banking
go run .
# listens on :8080
```

### 3. Make sure you have Ollama with qwen3

```bash
ollama pull qwen3:0.6b
```

## Examples

### 1) Structured mode (like v1, without intent LLM)

```bash
curl -X POST http://localhost:8080/ask \
  -H "Content-Type: application/json" \
  -d '{
    "operation": "get_balance",
    "params": { "accountId": "1234567890" }
  }'
```

### 2) Natural language mode (Intent Engine)

```bash
curl -X POST http://localhost:8080/ask_nlp \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What balance do I have in account 1234567890?"
  }'
```

```bash
curl -X POST http://localhost:8080/ask_nlp \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Show me the January transactions for account 1234567890"
  }'
```

```bash
curl -X POST http://localhost:8080/ask_nlp \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Send a 25-euro Bizum to my sister Laura at 600111222"
  }'
```

The response will include:

* `raw` → results from each tool in the pipeline (core, aml, bizum, notif…)
* `summary` → text from the Analyst LLM explaining what happened.

This v2 is a very solid foundation to extend to more domains and more banking operations.

## Secrets for tools (API keys, tokens)

When your YAML-defined tools need to call secured APIs, never hardcode secrets in the YAML. Instead, keep secrets in environment variables and reference them from your tool templates.

How it works:
- Tool templates (URL, body, and headers) support the template function env "VAR_NAME" to read an environment variable at runtime.
- Tools can declare HTTP headers in YAML under headers: (optional).

Example YAML tool with a bearer token from environment:

```yaml
tools:
  - name: crm.get_customer
    type: http
    method: GET
    url: "https://api.example.com/customers/{{ .customerId }}"
    timeout: 5000
    headers:
      Authorization: "Bearer {{ env \"CRM_API_TOKEN\" }}"
```

Notes and best practices:
- Set environment variables with your process manager or shell before starting AOS, e.g.:
  - macOS/Linux: export CRM_API_TOKEN="..."
  - systemd: Environment=CRM_API_TOKEN=... in the unit file
  - Docker: pass via -e CRM_API_TOKEN=...
- Do not commit secrets to Git. Keep them out of YAML files; reference them via env() instead.
- You can combine env() with normal parameter templates. For instance, custom headers or query parameters can interpolate both.
- Default Content-Type is application/json unless you override it in headers:.

Security consideration:
- env() simply reads the process environment. Prefer your platform’s secret store (Kubernetes Secrets, AWS/GCP/Azure Secret Manager, Docker Swarm/Compose secrets, etc.), mounted as environment variables at runtime.
