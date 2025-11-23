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

   - "Muéstrame el perfil del cliente 123"
   - "¿Qué interacciones recientes tiene el cliente 456 en los últimos 30 días?"
   - "Crea un ticket urgente para el cliente 789 porque tiene un problema con la factura"
