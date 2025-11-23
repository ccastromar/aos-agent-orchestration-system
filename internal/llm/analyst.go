package llm

import (
	"encoding/json"
	"fmt"
)

func SummarizeBankingResult(c LLMClient, intentType string, rawResult map[string]any) (string, error) {
	rawJSON, _ := json.Marshal(rawResult)

	prompt := fmt.Sprintf(`
Eres un asistente bancario experto.

Has ejecutado una operación con intent: "%s".
Aquí tienes los resultados en bruto de las herramientas (JSON):

%s

Escribe un resumen corto en español para el usuario final, explicando:
- qué operación se ha realizado,
- si todo ha ido bien,
- si el riesgo AML es bajo/medio/alto,
- cualquier detalle relevante (importe, saldo, movimientos...).

Devuelve SOLO texto plano, sin JSON, sin listas.
`, intentType, string(rawJSON))

	out, err := c.Chat(prompt)
	if err != nil {
		return "", err
	}
	return out, nil
}
