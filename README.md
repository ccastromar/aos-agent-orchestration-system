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
