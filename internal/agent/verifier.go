package agent

import (
	"fmt"
	"log"
	"time"

	"github.com/ccastromar/aos-banking-v2/internal/bus"
	"github.com/ccastromar/aos-banking-v2/internal/config"
	"github.com/ccastromar/aos-banking-v2/internal/logx"
	"github.com/ccastromar/aos-banking-v2/internal/tools"
	"github.com/ccastromar/aos-banking-v2/internal/ui"
)

type Verifier struct {
	bus     *bus.Bus
	cfg     *config.Config
	inbox   chan bus.Message
	uiStore *ui.UIStore
}

func NewVerifier(b *bus.Bus, cfg *config.Config, ui *ui.UIStore) *Verifier {
	return &Verifier{
		bus:     b,
		cfg:     cfg,
		inbox:   make(chan bus.Message, 16),
		uiStore: ui,
	}
}

func (v *Verifier) Inbox() chan bus.Message {
	return v.inbox
}

func (v *Verifier) Start() {
	for msg := range v.inbox {
		switch msg.Type {
		case "run_pipeline":
			v.handleRunPipeline(msg)
		default:
			log.Printf("[Verifier] mensaje desconocido: %#v", msg)
		}
	}
}

// -------------------------------------------------------------
// HANDLER PRINCIPAL
// -------------------------------------------------------------
func (v *Verifier) handleRunPipeline(msg bus.Message) {
	id := msg.Payload["id"].(string)
	intentType, _ := msg.Payload["intent"].(string)
	pipeAny := msg.Payload["pipeline"]
	paramsAny := msg.Payload["params"]

	pipe, ok := pipeAny.(config.Pipeline)
	if !ok {
		storeResult(id, Result{
			Status: "error",
			Err:    "pipeline inválido",
		})
		return
	}

	// Parámetros extraídos por el Planner
	baseParams := make(map[string]string)
	if paramsAny != nil {
		if mp, ok := paramsAny.(map[string]string); ok {
			for k, vv := range mp {
				baseParams[k] = vv
			}
		} else if mp2, ok := paramsAny.(map[string]any); ok {
			// Por compatibilidad si la decodificación viene en any
			for k, vv := range mp2 {
				if sv, ok := vv.(string); ok {
					baseParams[k] = sv
				}
			}
		}
	}

	log.Printf("[Verifier] ejecutando pipeline=%s id=%s intent=%s params=%#v",
		pipe.Name, id, intentType, baseParams)

	stepResults := make(map[string]any)

	// ---------------------------------------------------------
	// EJECUTAMOS CADA PASO
	// ---------------------------------------------------------
	for _, step := range pipe.Steps {

		// Paso ANALYST → directo al Analyst
		if step.Analyst {
			log.Printf("[Verifier] paso analyst=true id=%s -> llamando Analyst", id)
			v.bus.Send("analyst", bus.Message{
				Type: "summarize",
				Payload: map[string]any{
					"id":        id,
					"intent":    intentType,
					"rawResult": stepResults,
				},
			})
			return
		}

		// Paso TOOL
		toolName := step.Tool
		t, ok := v.cfg.Tools[toolName]
		if !ok {
			storeResult(id, Result{
				Status: "error",
				Err:    fmt.Sprintf("tool %s no encontrada", toolName),
			})
			return
		}

		log.Printf("[Verifier] ejecutando tool=%s id=%s", toolName, id)

		// Combinar parámetros → baseParams + WithParams
		callParams := make(map[string]string)
		for k, v := range baseParams {
			callParams[k] = v
		}
		for k, v := range step.WithParams {
			callParams[k] = v
		}

		// -----------------------------------------------------
		// RENDER TEMPLATE en Body
		// -----------------------------------------------------
		renderedBody, err := RenderTemplate(t.Body, callParams)
		if err != nil {
			log.Printf("[Verifier] error renderizando template tool=%s: %v", toolName, err)
			storeResult(id, Result{
				Status: "error",
				Err:    err.Error(),
			})
			return
		}

		// Ejecutar tool con cuerpo renderizado
		timer := logx.Start(id, "Verifier", "tool_"+toolName)
		start := time.Now()

		out, err := tools.ExecuteTool(t, renderedBody)
		timer.End()
		duration := time.Since(start).String()

		v.uiStore.AddEvent(id, "Verifier", "tool "+t.Name, "ok", duration)

		if err != nil {
			log.Printf("[Verifier] error ejecutando tool=%s: %v", toolName, err)
			storeResult(id, Result{
				Status: "error",
				Err:    err.Error(),
			})
			return
		}

		stepResults[toolName] = out
	}

	// Si terminó sin paso analyst explícito, lo enviamos ahora
	v.bus.Send("analyst", bus.Message{
		Type: "summarize",
		Payload: map[string]any{
			"id":        id,
			"intent":    intentType,
			"rawResult": stepResults,
		},
	})
}
