package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ccastromar/aos-banking-v2/internal/config"
)

// RenderTemplate aplica los parámetros a una plantilla Go.
// func RenderTemplate(tpl string, params map[string]string) (string, error) {
// 	if params == nil {
// 		return tpl, nil
// 	}

// 	t, err := template.New("tpl").Option("missingkey=default").Parse(tpl)
// 	if err != nil {
// 		return "", fmt.Errorf("error parseando template: %w", err)
// 	}

// 	var buf bytes.Buffer
// 	if err := t.Execute(&buf, params); err != nil {
// 		return "", fmt.Errorf("error ejecutando template: %w", err)
// 	}

// 	return buf.String(), nil
// }

// ExecuteTool ejecuta una tool HTTP, renderizando la URL y el body con parámetros.
func ExecuteTool(t config.Tool, params map[string]string) (map[string]any, error) {

	// 🔥 1. Renderizar la URL
	finalURL, err := RenderTemplateString(t.URL, params)
	if err != nil {
		return nil, fmt.Errorf("error renderizando URL: %w", err)
	}

	// 🔥 2. Renderizar el body
	bodyParams := map[string]string{}
	for k, v := range t.Body {
		rendered, err := RenderTemplateString(v, params)
		if err != nil {
			return nil, fmt.Errorf("error renderizando body: %w", err)
		}
		bodyParams[k] = rendered
	}

	// 3. Serializar body
	var payload []byte
	if len(bodyParams) > 0 {
		payload, err = json.Marshal(bodyParams)
		if err != nil {
			return nil, fmt.Errorf("error serializando body JSON: %w", err)
		}
	} else {
		payload = []byte("{}")
	}

	log.Printf("[Execute][DEBUG] finalURL=%s", finalURL)

	// 4. Crear request
	req, err := http.NewRequest(t.Method, finalURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 5. Enviar request
	client := &http.Client{
		Timeout: time.Duration(t.TimeoutMs) * time.Millisecond,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// 6. Leer respuesta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("[HTTP %d] %s", resp.StatusCode, string(respBody))
	}

	// 7. Parsear JSON
	out := map[string]any{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &out); err != nil {
			return nil, fmt.Errorf("error parseando JSON respuesta: %w", err)
		}
	}

	return out, nil
}
