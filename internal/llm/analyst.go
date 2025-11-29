package llm

import (
	"encoding/json"
	"fmt"
)

func SummarizeResult(c LLMClient, intentType string, rawResult map[string]any) (string, error) {
	rawJSON, _ := json.Marshal(rawResult)

	prompt := fmt.Sprintf(`
Eres un asistente multi dominio (banking, devops, CRM, Helpdesk, salud) experto.

Has ejecutado una operación con intent: "%s".
Aquí tienes los resultados en bruto de las herramientas (JSON):

%s

Escribe un resumen corto en español para el usuario final, explicando:
- qué operación se ha realizado,
- si todo ha ido bien,
- cualquier detalle relevante.

Devuelve SOLO texto plano, sin JSON, sin listas.
`, intentType, string(rawJSON))

	out, err := c.Chat(prompt)
	if err != nil {
		return "", err
	}
	return out, nil
}
