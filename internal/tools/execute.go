package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ccastromar/aos-banking-v2/internal/config"
)

// ExecuteTool ejecuta una tool HTTP usando el body ya renderizado.
// callParams ya NO se usa para templating; solo se mantiene para compatibilidad.
func ExecuteTool(t config.Tool, renderedBody map[string]string) (map[string]any, error) {

	// Convertir el mapa renderizado a JSON
	var payload []byte
	var err error

	if renderedBody != nil {
		payload, err = json.Marshal(renderedBody)
		if err != nil {
			return nil, fmt.Errorf("error serializando body: %w", err)
		}
	} else {
		payload = []byte("{}")
	}

	// Crear request
	req, err := http.NewRequest(t.Method, t.URL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Timeout configurable
	client := &http.Client{
		Timeout: time.Duration(t.TimeoutMs) * time.Millisecond,
	}

	// Ejecutar request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Leer respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var out map[string]any
	if len(body) > 0 {
		if err := json.Unmarshal(body, &out); err != nil {
			return nil, fmt.Errorf("error parseando JSON: %w", err)
		}
	}

	if out == nil {
		out = make(map[string]any)
	}

	return out, nil
}
