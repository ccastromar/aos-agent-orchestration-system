package llm

import (
	"encoding/json"
	"fmt"
	"strings"
)

type DetectedIntent struct {
	Type           string   `json:"intent"`
	RequiredParams []string `json:"required_params"`
}

type IntentSchema struct {
	Description string   `json:"description"`
	Params      []string `json:"params"`
}

// DetectIntent recibe el mensaje usuario + todos los intents del YAML
func DetectIntent(client Client, userMsg string, intents map[string]IntentSchema) (*DetectedIntent, error) {

	// preparar JSON para el prompt (el LLM verá todos los intents disponibles)
	intentsJSON, _ := json.Marshal(intents)

	prompt := fmt.Sprintf(`
Eres un clasificador estricto de intents bancarios.

Lista de intents disponibles:
%s

Devuelve EXCLUSIVAMENTE en JSON el intent que mejor coincide con el mensaje del usuario
más la lista de parámetros necesarios.

Formato de respuesta:
{
  "intent": "banking.get_balance",
  "required_params": ["accountId"]
}

Mensaje del usuario:
"%s"

Solo responde con JSON:
`, intentsJSON, userMsg)

	raw, err := client.Chat(prompt)
	if err != nil {
		return nil, err
	}

	raw = strings.TrimSpace(raw)

	var out DetectedIntent
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("DetectIntent JSON inválido: %w; raw=%s", err, raw)
	}

	// sanity check
	if out.Type == "" {
		return nil, fmt.Errorf("DetectIntent: intent vacío; raw=%s", raw)
	}

	return &out, nil
}
