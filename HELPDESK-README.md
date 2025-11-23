# AOS CRM Domain (Mock)

This adds a simple CRM domain to AOS with:

- Tools:
  - `crm_get_profile`
  - `crm_get_interactions`
  - `crm_create_ticket`
  - `crm_update_lead_status`
- Pipelines:
  - `pipeline_crm_profile`
  - `pipeline_crm_interactions`
  - `pipeline_crm_create_ticket`
  - `pipeline_crm_update_lead_status`
- Intents:
  - `crm.get_customer_profile`
  - `crm.get_customer_interactions`
  - `crm.create_ticket`
  - `crm.update_lead_status`
- Mock server:
  - Runs on `:9002`
  - Endpoints under `/mock/crm/*`

To use:

1. Start the mock CRM server:

   ```bash
   go run ./cmd/mock-crm
   ```

2. Ensure your AOS config loader points at `config/` (it will load `crm.yml` automatically).
3. Ask things like:

   curl -X POST http://localhost:8080/ask \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Crea un ticket urgente porque la web no funciona nada bien"
  }' | jq

curl -X POST http://localhost:8080/ask \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Muéstrame el estado del ticket TCK-ABC123"
  }' | jq

curl -X POST http://localhost:8080/ask \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Añade una nota al ticket TCK-999222 diciendo que ya revisé los logs"
  }' | jq

curl -X POST http://localhost:8080/ask \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Cierra el ticket TCK-123456 porque el usuario dice que ya está resuelto"
  }' | jq

curl -X POST http://localhost:8080/ask \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Tengo un cliente cabreadísimo, ábrele un ticket urgentísimo explicando que no puede pagar"
  }' | jq
