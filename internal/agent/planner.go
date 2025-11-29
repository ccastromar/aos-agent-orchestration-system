package agent

import (
	"context"

	"github.com/ccastromar/aos-agent-orchestration-system/internal/bus"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/config"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/guard"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/llm"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/logx"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/ui"
)

type Planner struct {
	bus       *bus.Bus
	cfg       *config.Config
	inbox     chan bus.Message
	llmClient llm.LLMClient
	uiStore   *ui.UIStore
}

func NewPlanner(b *bus.Bus, cfg *config.Config, llmClient llm.LLMClient, ui *ui.UIStore) *Planner {
	return &Planner{
		bus:       b,
		cfg:       cfg,
		inbox:     make(chan bus.Message, 16),
		llmClient: llmClient,
		uiStore:   ui,
	}
}

func (p *Planner) Inbox() chan bus.Message {
	return p.inbox
}
func (p *Planner) Start(ctx context.Context) error {
    defer func() {
        if r := recover(); r != nil {
            logx.Error("Planner", "panic recovered in Start: %v", r)
        }
    }()
    for {
        select {
        case msg := <-p.inbox:
            func() {
                defer func() {
                    if r := recover(); r != nil {
                        logx.Error("Planner", "panic recovered in dispatch: %v", r)
                    }
                }()
                p.dispatch(msg)
            }()

		case <-ctx.Done():
			return nil
		}
	}
}

func (p *Planner) dispatch(msg bus.Message) {
	switch msg.Type {
	case "detect_intent":
		p.handleDetectIntent(msg)

	case "new_task":
		// Esto viene del Inspector
		id := msg.Payload["id"].(string)
		logx.Debug("Planner", "got new_task id=%s -> forward to detect_intent", id)

		p.bus.Send("planner", bus.Message{
			Type: "detect_intent",
			Payload: map[string]any{
				"id":      id,
				"message": msg.Payload["message"],
				"mode":    msg.Payload["mode"],
			},
		})

	default:
		logx.Warn("Planner", "unknown message: %#v", msg)
	}

}

func (p *Planner) handleDetectIntent(msg bus.Message) {
    id := msg.Payload["id"].(string)
    userMsg := msg.Payload["message"].(string)

    logx.Debug("Planner", "detect_intent id=%s msg='%s'", id, userMsg)

    // obtain task context if present
    taskCtx, _ := GetTaskContext(id)
    if taskCtx == nil {
        taskCtx = context.Background()
    }

	intentKeys := make(map[string]any)
	for k := range p.cfg.Intents {
		intentKeys[k] = true
	}

 timer := logx.Start(id, "Planner", "DetectIntentLLM")
 detected, err := llm.DetectIntent(taskCtx, p.llmClient, userMsg, intentKeys)
 timer.End()
	if err != nil {
		logx.Error("Planner", "[%s] ERROR detecting intent: %v", id, err)
		//p.ui.AddEvent(id, "Planner", "intent_error", err.Error(), timer.Duration())
		storeResult(id, Result{Status: "error", Err: err.Error()})
		return
	}

	logx.Debug("Planner", "raw intent LLM='%s'", detected.Type)

	intentCfg, ok := p.cfg.Intents[detected.Type]
	if !ok {
		p.storeError(id, "intent desconocido para AOS")
		return
	}

	// 🔥 2. Seleccionar pipeline
	pipeName := intentCfg.Pipeline
	pipe, ok := p.cfg.Pipelines[pipeName]
	if !ok {
		p.storeError(id, "pipeline inexistente para intent")
		return
	}

	params := map[string]string{}

	// Traemos los required params del YAML
	required := intentCfg.RequiredParams

	if len(required) > 0 {
  timer := logx.Start(id, "Planner", "ExtractParams")
  extracted, err := llm.ExtractParams(taskCtx, p.llmClient, userMsg, required)
  timer.End()

		if err != nil {
			logx.Error("Planner", "[%s] ERROR extracting params: %v", id, err)
			p.storeError(id, "error extrayendo parámetros")
			return
		}

		params = extracted
	}

	err = guard.ValidateAll(intentCfg, pipe, params, p.cfg.Tools)
	if err != nil {
		logx.L(id, "Guard", "validation failed: %v", err)
		storeResult(id, Result{
			Status: "error",
			Err:    err.Error(),
		})
		return
	}

	logx.Info("Planner", "id=%s intent=%s pipeline=%s params=%v",
		id, detected.Type, pipeName, params)
	p.uiStore.AddEvent(id, "Planner", "intent", detected.Type, "")

	timer2 := logx.Start(id, "Planner", "DispatchPipeline")

	// 🔥 4. Enviar al Verifier
	p.bus.Send("verifier", bus.Message{
		Type: "run_pipeline",
		Payload: map[string]any{
			"id":       id,
			"intent":   detected.Type,
			"pipeline": pipe,
			"params":   params,
		},
	})
	timer2.End()

}

func (p *Planner) storeError(id string, errMsg string) {
	storeResult(id, Result{
		Status: "error",
		Err:    errMsg,
	})
}
