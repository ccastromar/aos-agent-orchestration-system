package agent

import (
	"log"

	"github.com/ccastromar/aos-banking-v2/internal/bus"
	"github.com/ccastromar/aos-banking-v2/internal/config"
	"github.com/ccastromar/aos-banking-v2/internal/guard"
	"github.com/ccastromar/aos-banking-v2/internal/llm"
	"github.com/ccastromar/aos-banking-v2/internal/logx"
	"github.com/ccastromar/aos-banking-v2/internal/ui"
)

type Planner struct {
	bus       *bus.Bus
	cfg       *config.Config
	inbox     chan bus.Message
	llmClient llm.Client
	uiStore   *ui.UIStore
}

func NewPlanner(b *bus.Bus, cfg *config.Config, llmClient llm.Client, ui *ui.UIStore) *Planner {
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

func (p *Planner) Start() {
	for msg := range p.inbox {
		switch msg.Type {
		case "detect_intent":
			p.handleDetectIntent(msg)

		case "new_task":
			// Esto viene del Inspector
			id := msg.Payload["id"].(string)
			log.Printf("[Planner][DEBUG] recibi new_task id=%s -> redirigiendo a detect_intent", id)

			p.bus.Send("planner", bus.Message{
				Type: "detect_intent",
				Payload: map[string]any{
					"id":      id,
					"message": msg.Payload["message"],
					"mode":    msg.Payload["mode"],
				},
			})

		default:
			log.Printf("[Planner] mensaje desconocido: %#v", msg)
		}
	}
}

func (p *Planner) handleDetectIntent(msg bus.Message) {
	id := msg.Payload["id"].(string)
	userMsg := msg.Payload["message"].(string)

	log.Printf("[Planner][DEBUG] detect_intent id=%s msg='%s'", id, userMsg)

	// Convertimos config.Intents → schema para el LLM
	intentSchemas := make(map[string]llm.IntentSchema)
	for intentName, it := range p.cfg.Intents {
		intentSchemas[intentName] = llm.IntentSchema{
			Description: it.Description,
			Params:      it.RequiredParams,
		}
	}

	timer := logx.Start(id, "Planner", "DetectIntentLLM")

	// 🔥 1. Detectar intent con LLM
	detected, err := llm.DetectIntent(p.llmClient, userMsg, intentSchemas)
	timer.End()

	if err != nil {
		log.Printf("[Planner] ERROR detectando intent: %v", err)
		p.storeError(id, "no se pudo detectar intent")
		return
	}

	log.Printf("[Planner][DEBUG] intent bruto LLM='%s'", detected.Type)

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

	// 🔥 3. Extraer parámetros si el intent los requiere
	params := map[string]string{}
	if len(detected.RequiredParams) > 0 {
		timer := logx.Start(id, "Planner", "ExtractParams")
		extracted, err := llm.ExtractParams(p.llmClient, userMsg, detected.RequiredParams)
		timer.End()

		if err != nil {
			log.Printf("[Planner] ERROR extrayendo parámetros: %v", err)
			p.storeError(id, "error extrayendo parámetros")
			return
		} else {
			params = extracted
		}

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

	log.Printf("[Planner] id=%s intent=%s pipeline=%s params=%v",
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
