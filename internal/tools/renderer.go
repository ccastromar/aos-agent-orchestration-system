package tools

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
)

//
// -------------------------------------------------------------
// TEMPLATE RENDERING UTILITIES
// -------------------------------------------------------------
//

// RenderTemplateString procesa un template que es STRING.
// Sirve para URLs tipo:
//
//	"http://localhost:9000/svc?id={{ .id }}"
func RenderTemplateString(tpl string, params map[string]string) (string, error) {
	if params == nil {
		return tpl, nil
	}

	t, err := template.New("tpl").
		Option("missingkey=default").
		Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("error parseando template string: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("error ejecutando template string: %w", err)
	}

	return buf.String(), nil
}

// RenderTemplateMap procesa UN MAP de strings.
// Sirve para el body de las tools, por ejemplo:
//
// body:
//
//	customerId: "{{ .customerId }}"
//	days:       "{{ .days }}"
//
// Produce un map[string]string renderizado.
func RenderTemplateMap(body map[string]string, params map[string]string) (map[string]string, error) {
	if body == nil {
		return map[string]string{}, nil
	}

	out := make(map[string]string)

	for k, v := range body {
		t, err := template.New("body").
			Option("missingkey=default").
			Parse(v)
		if err != nil {
			return nil, fmt.Errorf("error parseando template body campo=%s: %w", k, err)
		}

		var buf bytes.Buffer
		if err := t.Execute(&buf, params); err != nil {
			return nil, fmt.Errorf("error ejecutando template body campo=%s: %w", k, err)
		}

		out[k] = buf.String()
	}

	return out, nil
}

//
// -------------------------------------------------------------
// DEBUG HELPERS
// -------------------------------------------------------------
//

// DebugRender muestra cómo queda un template string (para logging manual)
func DebugRender(label, tpl string, params map[string]string) {
	out, err := RenderTemplateString(tpl, params)
	if err != nil {
		log.Printf("[TEMPLATE][%s] ERROR: %v", label, err)
	} else {
		log.Printf("[TEMPLATE][%s] %s => %s", label, tpl, out)
	}
}
