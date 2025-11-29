package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/ccastromar/aos-agent-orchestration-system/internal/bus"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/config"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/logx"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/tools"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/ui"
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

func (v *Verifier) Start(ctx context.Context) error {
    defer func() {
        if r := recover(); r != nil {
            logx.Error("Verifier", "panic recovered in Start: %v", r)
        }
    }()
    for {
        select {
        case msg := <-v.inbox:
            func() {
                defer func() {
                    if r := recover(); r != nil {
                        logx.Error("Verifier", "panic recovered in dispatch: %v", r)
                    }
                }()
                v.dispatch(msg)
            }()

        case <-ctx.Done():
            return nil
        }
    }
}

func (v *Verifier) dispatch(msg bus.Message) {
	switch msg.Type {
	case "run_pipeline":
		v.handleRunPipeline(msg)
	default:
		logx.Warn("Verifier", "unknown message: %#v", msg)
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

	logx.Info("Verifier", "executing pipeline=%s id=%s intent=%s params=%#v",
		pipe.Name, id, intentType, baseParams)

	stepResults := make(map[string]any)

	// ---------------------------------------------------------
	// EJECUTAMOS CADA PASO
	// ---------------------------------------------------------
	for _, step := range pipe.Steps {

		// Step ANALYST → directo al Analyst
		if step.Analyst {
			logx.Debug("Verifier", "analyst=true id=%s -> calling Analyst", id)
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

		// Step TOOL
		toolName := step.Tool
		t, ok := v.cfg.Tools[toolName]
		if !ok {
			storeResult(id, Result{
				Status: "error",
				Err:    fmt.Sprintf("tool %s no encontrada", toolName),
			})
			return
		}

		logx.Info("Verifier", "executing tool=%s id=%s", toolName, id)
		// Combinar parámetros → baseParams + WithParams (sin pisar los del Planner)
		callParams := make(map[string]string)

		// 1. Copiar params del Planner
		for k, v := range baseParams {
			callParams[k] = v
		}

		// 2. Rellenar defaults del pipeline SIN sobreescribir valores existentes
		for k, v := range step.WithParams {
			if _, exists := callParams[k]; !exists || callParams[k] == "" {
				if v != "" { // evitamos meter valores vacíos
					callParams[k] = v
				}
			}
		}

		// Ejecutar tool con cuerpo renderizado
		timer := logx.Start(id, "Verifier", "tool_"+toolName)
		start := time.Now()

		logx.Debug("Verifier", "params for the tool=%s id=%s params=%#v",
			toolName, id, callParams)

		// obtain task context if present
		taskCtx, _ := GetTaskContext(id)
		if taskCtx == nil {
			taskCtx = context.Background()
		}

		out, err := tools.ExecuteToolCtx(taskCtx, t, callParams)
		timer.End()
		duration := time.Since(start).String()

		v.uiStore.AddEvent(id, "Verifier", "tool "+t.Name, "ok", duration)

		if err != nil {
			logx.Error("Verifier", "error executing tool=%s: %v", toolName, err)
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
