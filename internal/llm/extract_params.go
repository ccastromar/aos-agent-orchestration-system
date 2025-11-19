package llm

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ExtractParams solicita al LLM que extraiga SOLO los parámetros necesarios
// definidos en el YAML del intent.
func ExtractParams(client Client, userMessage string, required []string) (map[string]string, error) {

	if len(required) == 0 {
		return map[string]string{}, nil
	}

	prompt := fmt.Sprintf(`
Extrae exclusivamente los siguientes parámetros del mensaje del usuario:

%v

Devuelve ÚNICAMENTE un JSON plano con esos campos.
Si un parámetro no aparece en el mensaje, deja su valor vacío.

Ejemplo de formato:
{
  "amount": "20",
  "toPhone": "Laura",
  "concept": "regalo"
}

Mensaje del usuario:
"%s"

Solo responde con el JSON:
`, required, userMessage)

	raw, err := client.Chat(prompt)
	if err != nil {
		return nil, err
	}

	raw = strings.TrimSpace(raw)

	out := map[string]string{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("error parseando JSON de parámetros: %w; raw=%s", err, raw)
	}

	return out, nil
}
